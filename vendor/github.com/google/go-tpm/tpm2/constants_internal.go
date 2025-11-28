package tpm2

// This file contains constant definitions we don't want to use stringer with
// (because they are duplicates of other values, and we would prefer those values
// to influence the string representations).

// Hash algorithm IDs and command codes that got re-used.
const (
	TPMAlgSHA          = TPMAlgSHA1
	TPMCCHMAC          = TPMCCMAC
	TPMCCHMACStart     = TPMCCMACStart
	TPMHTLoadedSession = TPMHTHMACSession
	TPMHTSavedSession  = TPMHTPolicySession
)
