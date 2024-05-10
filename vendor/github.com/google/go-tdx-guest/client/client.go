// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package client provides the library functions to get a TDX quote
// from the TDX guest device
package client

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/go-tdx-guest/abi"
	labi "github.com/google/go-tdx-guest/client/linuxabi"
	"github.com/google/logger"
)

var tdxGuestPath = flag.String("tdx_guest_device_path", "default",
	"Path to TDX guest device. If \"default\", uses platform default or a fake if testing.")

// Device encapsulates the possible commands to the TDX guest device.
// Deprecated: The Device interface is deprecated, and use of quote provider interface is
// recommended for fetching attestation quote.
type Device interface {
	Open(path string) error
	Close() error
	Ioctl(command uintptr, argument any) (uintptr, error)
}

// QuoteProvider encapsulates calls to attestation quote.
type QuoteProvider interface {
	IsSupported() error
	GetRawQuote(reportData [64]byte) ([]uint8, error)
}

// UseDefaultTdxGuestDevice returns true if tdxGuestPath=default.
func UseDefaultTdxGuestDevice() bool {
	return *tdxGuestPath == "default"
}

// getReport requests for tdx report by making an ioctl call.
func getReport(d Device, reportData [64]byte) ([]uint8, error) {
	tdxReportReq := labi.TdxReportReq{}
	copy(tdxReportReq.ReportData[:], reportData[:])
	result, err := d.Ioctl(labi.IocTdxGetReport, &tdxReportReq)
	if err != nil {
		return nil, err
	}
	if result != uintptr(labi.TdxAttestSuccess) {
		return nil, fmt.Errorf("unable to get the report: %d", result)
	}
	return tdxReportReq.TdReport[:], nil
}

// GetRawQuote uses Quote provider or Device(deprecated) to get the quote in byte array.
func GetRawQuote(quoteProvider any, reportData [64]byte) ([]uint8, error) {
	switch qp := quoteProvider.(type) {
	case Device:
		return getRawQuoteViaDevice(qp, reportData)
	case QuoteProvider:
		return getRawQuoteViaProvider(qp, reportData)
	default:
		return nil, fmt.Errorf("unsupported quote provider type: %T", quoteProvider)
	}
}

// getRawQuoteViaDevice uses TDX device driver to call getReport for report and convert it to
// quote using an ioctl call.
func getRawQuoteViaDevice(d Device, reportData [64]byte) ([]uint8, error) {
	logger.V(1).Info("Get raw TDX quote via Device")
	tdReport, err := getReport(d, reportData)
	if err != nil {
		return nil, err
	}
	tdxHdr := &labi.TdxQuoteHdr{
		Status:  0,
		Version: 1,
		InLen:   labi.TdReportSize,
		OutLen:  0,
	}
	copy(tdxHdr.Data[:], tdReport[:labi.TdReportSize])
	tdxReq := labi.TdxQuoteReq{
		Buffer: tdxHdr,
		Length: labi.ReqBufSize,
	}
	result, err := d.Ioctl(labi.IocTdxGetQuote, &tdxReq)
	if err != nil {
		return nil, err
	}
	if result != uintptr(labi.TdxAttestSuccess) {
		return nil, fmt.Errorf("unable to get the quote")
	}
	if tdxHdr.Status != 0 {
		if labi.GetQuoteInFlight == tdxHdr.Status {
			return nil, fmt.Errorf("the device driver return busy")
		} else if labi.GetQuoteServiceUnavailable == tdxHdr.Status {
			return nil, fmt.Errorf("request feature is not supported")
		} else if tdxHdr.OutLen == 0 || tdxHdr.OutLen > labi.ReqBufSize {
			return nil, fmt.Errorf("invalid Quote size: %v. It must be > 0 and <= : %v", tdxHdr.OutLen, labi.ReqBufSize)
		}

		return nil, fmt.Errorf("unexpected error: %v", tdxHdr.Status)
	}

	return tdxHdr.Data[:tdxHdr.OutLen], nil
}

// getRawQuoteViaProvider use QuoteProvider to fetch quote in byte array format.
func getRawQuoteViaProvider(qp QuoteProvider, reportData [64]byte) ([]uint8, error) {
	if err := qp.IsSupported(); err == nil {
		logger.V(1).Info("Get raw TDX quote via QuoteProvider")
		quote, err := qp.GetRawQuote(reportData)
		return quote, err
	}
	return fallbackToDeviceForRawQuote(reportData)
}

// GetQuote uses Quote provider or Device(deprecated) to get the quote in byte array and convert it
// into proto.
// Supported quote formats - QuoteV4.
func GetQuote(quoteProvider any, reportData [64]byte) (any, error) {
	quotebytes, err := GetRawQuote(quoteProvider, reportData)
	if err != nil {
		return nil, err
	}
	quote, err := abi.QuoteToProto(quotebytes)
	if err != nil {
		return nil, err
	}
	return quote, nil
}

// fallbackToDeviceForRawQuote opens tdx_guest device to fetch raw quote.
func fallbackToDeviceForRawQuote(reportData [64]byte) ([]uint8, error) {
	// Fall back to TDX device driver.
	device, err := OpenDevice()
	if err != nil {
		return nil, fmt.Errorf("neither TDX device, nor ConfigFs is available to fetch attestation quote")
	}
	bytes, err := getRawQuoteViaDevice(device, reportData)
	device.Close()
	return bytes, err
}

func init() {
	logger.Init("", false, false, os.Stdout)
}
