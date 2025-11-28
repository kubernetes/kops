package tpm2

//go:generate stringer -trimprefix=TPM -type=TPMAlgID,TPMECCCurve,TPMCC,TPMRC,TPMEO,TPMST,TPMCap,TPMPT,TPMPTPCR,TPMHT,TPMHandle,TPMNT -output=constants_string.go constants.go

import (

	// Register the relevant hash implementations.
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
)

// TPMAlgID represents a TPM_ALG_ID.
// See definition in Part 2: Structures, section 6.3.
type TPMAlgID uint16

// TPMAlgID values come from Part 2: Structures, section 6.3.
const (
	TPMAlgRSA          TPMAlgID = 0x0001
	TPMAlgTDES         TPMAlgID = 0x0003
	TPMAlgSHA1         TPMAlgID = 0x0004
	TPMAlgHMAC         TPMAlgID = 0x0005
	TPMAlgAES          TPMAlgID = 0x0006
	TPMAlgMGF1         TPMAlgID = 0x0007
	TPMAlgKeyedHash    TPMAlgID = 0x0008
	TPMAlgXOR          TPMAlgID = 0x000A
	TPMAlgSHA256       TPMAlgID = 0x000B
	TPMAlgSHA384       TPMAlgID = 0x000C
	TPMAlgSHA512       TPMAlgID = 0x000D
	TPMAlgSHA256192    TPMAlgID = 0x000E
	TPMAlgNull         TPMAlgID = 0x0010
	TPMAlgSM3256       TPMAlgID = 0x0012
	TPMAlgSM4          TPMAlgID = 0x0013
	TPMAlgRSASSA       TPMAlgID = 0x0014
	TPMAlgRSAES        TPMAlgID = 0x0015
	TPMAlgRSAPSS       TPMAlgID = 0x0016
	TPMAlgOAEP         TPMAlgID = 0x0017
	TPMAlgECDSA        TPMAlgID = 0x0018
	TPMAlgECDH         TPMAlgID = 0x0019
	TPMAlgECDAA        TPMAlgID = 0x001A
	TPMAlgSM2          TPMAlgID = 0x001B
	TPMAlgECSchnorr    TPMAlgID = 0x001C
	TPMAlgECMQV        TPMAlgID = 0x001D
	TPMAlgKDF1SP80056A TPMAlgID = 0x0020
	TPMAlgKDF2         TPMAlgID = 0x0021
	TPMAlgKDF1SP800108 TPMAlgID = 0x0022
	TPMAlgECC          TPMAlgID = 0x0023
	TPMAlgSymCipher    TPMAlgID = 0x0025
	TPMAlgCamellia     TPMAlgID = 0x0026
	TPMAlgSHA3256      TPMAlgID = 0x0027
	TPMAlgSHA3384      TPMAlgID = 0x0028
	TPMAlgSHA3512      TPMAlgID = 0x0029
	TPMAlgSHAKE128     TPMAlgID = 0x002A
	TPMAlgSHAKE256     TPMAlgID = 0x002B
	TPMAlgSHAKE256192  TPMAlgID = 0x002C
	TPMAlgSHAKE256256  TPMAlgID = 0x002D
	TPMAlgSHAKE256512  TPMAlgID = 0x002E
	TPMAlgCMAC         TPMAlgID = 0x003F
	TPMAlgCTR          TPMAlgID = 0x0040
	TPMAlgOFB          TPMAlgID = 0x0041
	TPMAlgCBC          TPMAlgID = 0x0042
	TPMAlgCFB          TPMAlgID = 0x0043
	TPMAlgECB          TPMAlgID = 0x0044
	TPMAlgCCM          TPMAlgID = 0x0050
	TPMAlgGCM          TPMAlgID = 0x0051
	TPMAlgKW           TPMAlgID = 0x0052
	TPMAlgKWP          TPMAlgID = 0x0053
	TPMAlgEAX          TPMAlgID = 0x0054
	TPMAlgEDDSA        TPMAlgID = 0x0060
	TPMAlgEDDSAPH      TPMAlgID = 0x0061
	TPMAlgLMS          TPMAlgID = 0x0070
	TPMAlgXMSS         TPMAlgID = 0x0071
	TPMAlgKEYEDXOF     TPMAlgID = 0x0080
	TPMAlgKMACXOF128   TPMAlgID = 0x0081
	TPMAlgKMACXOF256   TPMAlgID = 0x0082
	TPMAlgKMAC128      TPMAlgID = 0x0090
	TPMAlgKMAC256      TPMAlgID = 0x0091
)

// TPMECCCurve represents a TPM_ECC_Curve.
// See definition in Part 2: Structures, section 6.4.
type TPMECCCurve uint16

// TPMECCCurve values come from Part 2: Structures, section 6.4.
const (
	TPMECCNone            TPMECCCurve = 0x0000
	TPMECCNistP192        TPMECCCurve = 0x0001
	TPMECCNistP224        TPMECCCurve = 0x0002
	TPMECCNistP256        TPMECCCurve = 0x0003
	TPMECCNistP384        TPMECCCurve = 0x0004
	TPMECCNistP521        TPMECCCurve = 0x0005
	TPMECCBNP256          TPMECCCurve = 0x0010
	TPMECCBNP638          TPMECCCurve = 0x0011
	TPMECCSM2P256         TPMECCCurve = 0x0020
	TPMECCBrainpoolP256R1 TPMECCCurve = 0x0030
	TPMECCBrainpoolP384R1 TPMECCCurve = 0x0031
	TPMECCBrainpoolP512R1 TPMECCCurve = 0x0032
	TPMECCCurve25519      TPMECCCurve = 0x0040
	TPMECCCurve448        TPMECCCurve = 0x0041
)

// TPMCC represents a TPM_CC.
// See definition in Part 2: Structures, section 6.5.2.
type TPMCC uint32

// TPMCC values come from Part 2: Structures, section 6.5.2.
const (
	TPMCCNVUndefineSpaceSpecial     TPMCC = 0x0000011F
	TPMCCEvictControl               TPMCC = 0x00000120
	TPMCCHierarchyControl           TPMCC = 0x00000121
	TPMCCNVUndefineSpace            TPMCC = 0x00000122
	TPMCCChangeEPS                  TPMCC = 0x00000124
	TPMCCChangePPS                  TPMCC = 0x00000125
	TPMCCClear                      TPMCC = 0x00000126
	TPMCCClearControl               TPMCC = 0x00000127
	TPMCCClockSet                   TPMCC = 0x00000128
	TPMCCHierarchyChanegAuth        TPMCC = 0x00000129
	TPMCCNVDefineSpace              TPMCC = 0x0000012A
	TPMCCPCRAllocate                TPMCC = 0x0000012B
	TPMCCPCRSetAuthPolicy           TPMCC = 0x0000012C
	TPMCCPPCommands                 TPMCC = 0x0000012D
	TPMCCSetPrimaryPolicy           TPMCC = 0x0000012E
	TPMCCFieldUpgradeStart          TPMCC = 0x0000012F
	TPMCCClockRateAdjust            TPMCC = 0x00000130
	TPMCCCreatePrimary              TPMCC = 0x00000131
	TPMCCNVGlobalWriteLock          TPMCC = 0x00000132
	TPMCCGetCommandAuditDigest      TPMCC = 0x00000133
	TPMCCNVIncrement                TPMCC = 0x00000134
	TPMCCNVSetBits                  TPMCC = 0x00000135
	TPMCCNVExtend                   TPMCC = 0x00000136
	TPMCCNVWrite                    TPMCC = 0x00000137
	TPMCCNVWriteLock                TPMCC = 0x00000138
	TPMCCDictionaryAttackLockReset  TPMCC = 0x00000139
	TPMCCDictionaryAttackParameters TPMCC = 0x0000013A
	TPMCCNVChangeAuth               TPMCC = 0x0000013B
	TPMCCPCREvent                   TPMCC = 0x0000013C
	TPMCCPCRReset                   TPMCC = 0x0000013D
	TPMCCSequenceComplete           TPMCC = 0x0000013E
	TPMCCSetAlgorithmSet            TPMCC = 0x0000013F
	TPMCCSetCommandCodeAuditStatus  TPMCC = 0x00000140
	TPMCCFieldUpgradeData           TPMCC = 0x00000141
	TPMCCIncrementalSelfTest        TPMCC = 0x00000142
	TPMCCSelfTest                   TPMCC = 0x00000143
	TPMCCStartup                    TPMCC = 0x00000144
	TPMCCShutdown                   TPMCC = 0x00000145
	TPMCCStirRandom                 TPMCC = 0x00000146
	TPMCCActivateCredential         TPMCC = 0x00000147
	TPMCCCertify                    TPMCC = 0x00000148
	TPMCCPolicyNV                   TPMCC = 0x00000149
	TPMCCCertifyCreation            TPMCC = 0x0000014A
	TPMCCDuplicate                  TPMCC = 0x0000014B
	TPMCCGetTime                    TPMCC = 0x0000014C
	TPMCCGetSessionAuditDigest      TPMCC = 0x0000014D
	TPMCCNVRead                     TPMCC = 0x0000014E
	TPMCCNVReadLock                 TPMCC = 0x0000014F
	TPMCCObjectChangeAuth           TPMCC = 0x00000150
	TPMCCPolicySecret               TPMCC = 0x00000151
	TPMCCRewrap                     TPMCC = 0x00000152
	TPMCCCreate                     TPMCC = 0x00000153
	TPMCCECDHZGen                   TPMCC = 0x00000154
	TPMCCMAC                        TPMCC = 0x00000155
	TPMCCImport                     TPMCC = 0x00000156
	TPMCCLoad                       TPMCC = 0x00000157
	TPMCCQuote                      TPMCC = 0x00000158
	TPMCCRSADecrypt                 TPMCC = 0x00000159
	TPMCCMACStart                   TPMCC = 0x0000015B
	TPMCCSequenceUpdate             TPMCC = 0x0000015C
	TPMCCSign                       TPMCC = 0x0000015D
	TPMCCUnseal                     TPMCC = 0x0000015E
	TPMCCPolicySigned               TPMCC = 0x00000160
	TPMCCContextLoad                TPMCC = 0x00000161
	TPMCCContextSave                TPMCC = 0x00000162
	TPMCCECDHKeyGen                 TPMCC = 0x00000163
	TPMCCEncryptDecrypt             TPMCC = 0x00000164
	TPMCCFlushContext               TPMCC = 0x00000165
	TPMCCLoadExternal               TPMCC = 0x00000167
	TPMCCMakeCredential             TPMCC = 0x00000168
	TPMCCNVReadPublic               TPMCC = 0x00000169
	TPMCCPolicyAuthorize            TPMCC = 0x0000016A
	TPMCCPolicyAuthValue            TPMCC = 0x0000016B
	TPMCCPolicyCommandCode          TPMCC = 0x0000016C
	TPMCCPolicyCounterTimer         TPMCC = 0x0000016D
	TPMCCPolicyCpHash               TPMCC = 0x0000016E
	TPMCCPolicyLocality             TPMCC = 0x0000016F
	TPMCCPolicyNameHash             TPMCC = 0x00000170
	TPMCCPolicyOR                   TPMCC = 0x00000171
	TPMCCPolicyTicket               TPMCC = 0x00000172
	TPMCCReadPublic                 TPMCC = 0x00000173
	TPMCCRSAEncrypt                 TPMCC = 0x00000174
	TPMCCStartAuthSession           TPMCC = 0x00000176
	TPMCCVerifySignature            TPMCC = 0x00000177
	TPMCCECCParameters              TPMCC = 0x00000178
	TPMCCFirmwareRead               TPMCC = 0x00000179
	TPMCCGetCapability              TPMCC = 0x0000017A
	TPMCCGetRandom                  TPMCC = 0x0000017B
	TPMCCGetTestResult              TPMCC = 0x0000017C
	TPMCCHash                       TPMCC = 0x0000017D
	TPMCCPCRRead                    TPMCC = 0x0000017E
	TPMCCPolicyPCR                  TPMCC = 0x0000017F
	TPMCCPolicyRestart              TPMCC = 0x00000180
	TPMCCReadClock                  TPMCC = 0x00000181
	TPMCCPCRExtend                  TPMCC = 0x00000182
	TPMCCPCRSetAuthValue            TPMCC = 0x00000183
	TPMCCNVCertify                  TPMCC = 0x00000184
	TPMCCEventSequenceComplete      TPMCC = 0x00000185
	TPMCCHashSequenceStart          TPMCC = 0x00000186
	TPMCCPolicyPhysicalPresence     TPMCC = 0x00000187
	TPMCCPolicyDuplicationSelect    TPMCC = 0x00000188
	TPMCCPolicyGetDigest            TPMCC = 0x00000189
	TPMCCTestParms                  TPMCC = 0x0000018A
	TPMCCCommit                     TPMCC = 0x0000018B
	TPMCCPolicyPassword             TPMCC = 0x0000018C
	TPMCCZGen2Phase                 TPMCC = 0x0000018D
	TPMCCECEphemeral                TPMCC = 0x0000018E
	TPMCCPolicyNvWritten            TPMCC = 0x0000018F
	TPMCCPolicyTemplate             TPMCC = 0x00000190
	TPMCCCreateLoaded               TPMCC = 0x00000191
	TPMCCPolicyAuthorizeNV          TPMCC = 0x00000192
	TPMCCEncryptDecrypt2            TPMCC = 0x00000193
	TPMCCACGetCapability            TPMCC = 0x00000194
	TPMCCACSend                     TPMCC = 0x00000195
	TPMCCPolicyACSendSelect         TPMCC = 0x00000196
	TPMCCCertifyX509                TPMCC = 0x00000197
	TPMCCACTSetTimeout              TPMCC = 0x00000198
)

// TPMRC represents a TPM_RC.
// See definition in Part 2: Structures, section 6.6.
type TPMRC uint32

// TPMRC values come from Part 2: Structures, section 6.6.3.
const (
	rcVer1             = 0x00000100
	rcFmt1             = 0x00000080
	rcWarn             = 0x00000900
	rcP                = 0x00000040
	rcS                = 0x00000800
	TPMRCSuccess TPMRC = 0x00000000
	// FMT0 error codes
	TPMRCInitialize      TPMRC = rcVer1 + 0x000
	TPMRCFailure         TPMRC = rcVer1 + 0x001
	TPMRCSequence        TPMRC = rcVer1 + 0x003
	TPMRCPrivate         TPMRC = rcVer1 + 0x00B
	TPMRCHMAC            TPMRC = rcVer1 + 0x019
	TPMRCDisabled        TPMRC = rcVer1 + 0x020
	TPMRCExclusive       TPMRC = rcVer1 + 0x021
	TPMRCAuthType        TPMRC = rcVer1 + 0x024
	TPMRCAuthMissing     TPMRC = rcVer1 + 0x025
	TPMRCPolicy          TPMRC = rcVer1 + 0x026
	TPMRCPCR             TPMRC = rcVer1 + 0x027
	TPMRCPCRChanged      TPMRC = rcVer1 + 0x028
	TPMRCUpgrade         TPMRC = rcVer1 + 0x02D
	TPMRCTooManyContexts TPMRC = rcVer1 + 0x02E
	TPMRCAuthUnavailable TPMRC = rcVer1 + 0x02F
	TPMRCReboot          TPMRC = rcVer1 + 0x030
	TPMRCUnbalanced      TPMRC = rcVer1 + 0x031
	TPMRCCommandSize     TPMRC = rcVer1 + 0x042
	TPMRCCommandCode     TPMRC = rcVer1 + 0x043
	TPMRCAuthSize        TPMRC = rcVer1 + 0x044
	TPMRCAuthContext     TPMRC = rcVer1 + 0x045
	TPMRCNVRange         TPMRC = rcVer1 + 0x046
	TPMRCNVSize          TPMRC = rcVer1 + 0x047
	TPMRCNVLocked        TPMRC = rcVer1 + 0x048
	TPMRCNVAuthorization TPMRC = rcVer1 + 0x049
	TPMRCNVUninitialized TPMRC = rcVer1 + 0x04A
	TPMRCNVSpace         TPMRC = rcVer1 + 0x04B
	TPMRCNVDefined       TPMRC = rcVer1 + 0x04C
	TPMRCBadContext      TPMRC = rcVer1 + 0x050
	TPMRCCPHash          TPMRC = rcVer1 + 0x051
	TPMRCParent          TPMRC = rcVer1 + 0x052
	TPMRCNeedsTest       TPMRC = rcVer1 + 0x053
	TPMRCNoResult        TPMRC = rcVer1 + 0x054
	TPMRCSensitive       TPMRC = rcVer1 + 0x055
	// FMT1 error codes
	TPMRCAsymmetric   TPMRC = rcFmt1 + 0x001
	TPMRCAttributes   TPMRC = rcFmt1 + 0x002
	TPMRCHash         TPMRC = rcFmt1 + 0x003
	TPMRCValue        TPMRC = rcFmt1 + 0x004
	TPMRCHierarchy    TPMRC = rcFmt1 + 0x005
	TPMRCKeySize      TPMRC = rcFmt1 + 0x007
	TPMRCMGF          TPMRC = rcFmt1 + 0x008
	TPMRCMode         TPMRC = rcFmt1 + 0x009
	TPMRCType         TPMRC = rcFmt1 + 0x00A
	TPMRCHandle       TPMRC = rcFmt1 + 0x00B
	TPMRCKDF          TPMRC = rcFmt1 + 0x00C
	TPMRCRange        TPMRC = rcFmt1 + 0x00D
	TPMRCAuthFail     TPMRC = rcFmt1 + 0x00E
	TPMRCNonce        TPMRC = rcFmt1 + 0x00F
	TPMRCPP           TPMRC = rcFmt1 + 0x010
	TPMRCScheme       TPMRC = rcFmt1 + 0x012
	TPMRCSize         TPMRC = rcFmt1 + 0x015
	TPMRCSymmetric    TPMRC = rcFmt1 + 0x016
	TPMRCTag          TPMRC = rcFmt1 + 0x017
	TPMRCSelector     TPMRC = rcFmt1 + 0x018
	TPMRCInsufficient TPMRC = rcFmt1 + 0x01A
	TPMRCSignature    TPMRC = rcFmt1 + 0x01B
	TPMRCKey          TPMRC = rcFmt1 + 0x01C
	TPMRCPolicyFail   TPMRC = rcFmt1 + 0x01D
	TPMRCIntegrity    TPMRC = rcFmt1 + 0x01F
	TPMRCTicket       TPMRC = rcFmt1 + 0x020
	TPMRCReservedBits TPMRC = rcFmt1 + 0x021
	TPMRCBadAuth      TPMRC = rcFmt1 + 0x022
	TPMRCExpired      TPMRC = rcFmt1 + 0x023
	TPMRCPolicyCC     TPMRC = rcFmt1 + 0x024
	TPMRCBinding      TPMRC = rcFmt1 + 0x025
	TPMRCCurve        TPMRC = rcFmt1 + 0x026
	TPMRCECCPoint     TPMRC = rcFmt1 + 0x027
	// Warnings
	TPMRCContextGap     TPMRC = rcWarn + 0x001
	TPMRCObjectMemory   TPMRC = rcWarn + 0x002
	TPMRCSessionMemory  TPMRC = rcWarn + 0x003
	TPMRCMemory         TPMRC = rcWarn + 0x004
	TPMRCSessionHandles TPMRC = rcWarn + 0x005
	TPMRCObjectHandles  TPMRC = rcWarn + 0x006
	TPMRCLocality       TPMRC = rcWarn + 0x007
	TPMRCYielded        TPMRC = rcWarn + 0x008
	TPMRCCanceled       TPMRC = rcWarn + 0x009
	TPMRCTesting        TPMRC = rcWarn + 0x00A
	TPMRCReferenceH0    TPMRC = rcWarn + 0x010
	TPMRCReferenceH1    TPMRC = rcWarn + 0x011
	TPMRCReferenceH2    TPMRC = rcWarn + 0x012
	TPMRCReferenceH3    TPMRC = rcWarn + 0x013
	TPMRCReferenceH4    TPMRC = rcWarn + 0x014
	TPMRCReferenceH5    TPMRC = rcWarn + 0x015
	TPMRCReferenceH6    TPMRC = rcWarn + 0x016
	TPMRCReferenceS0    TPMRC = rcWarn + 0x018
	TPMRCReferenceS1    TPMRC = rcWarn + 0x019
	TPMRCReferenceS2    TPMRC = rcWarn + 0x01A
	TPMRCReferenceS3    TPMRC = rcWarn + 0x01B
	TPMRCReferenceS4    TPMRC = rcWarn + 0x01C
	TPMRCReferenceS5    TPMRC = rcWarn + 0x01D
	TPMRCReferenceS6    TPMRC = rcWarn + 0x01E
	TPMRCNVRate         TPMRC = rcWarn + 0x020
	TPMRCLockout        TPMRC = rcWarn + 0x021
	TPMRCRetry          TPMRC = rcWarn + 0x022
	TPMRCNVUnavailable  TPMRC = rcWarn + 0x023
)

// TPMEO represents a TPM_EO.
// See definition in Part 2: Structures, section 6.8.
type TPMEO uint16

// TPMEO values come from Part 2: Structures, section 6.8.
const (
	TPMEOEq         TPMEO = 0x0000
	TPMEONeq        TPMEO = 0x0001
	TPMEOSignedGT   TPMEO = 0x0002
	TPMEOUnsignedGT TPMEO = 0x0003
	TPMEOSignedLT   TPMEO = 0x0004
	TPMEOUnsignedLT TPMEO = 0x0005
	TPMEOSignedGE   TPMEO = 0x0006
	TPMEOUnsignedGE TPMEO = 0x0007
	TPMEOSignedLE   TPMEO = 0x0008
	TPMEOUnsignedLE TPMEO = 0x0009
	TPMEOBitSet     TPMEO = 0x000A
	TPMEOBitClear   TPMEO = 0x000B
)

// TPMST represents a TPM_ST.
// See definition in Part 2: Structures, section 6.9.
type TPMST uint16

// TPMST values come from Part 2: Structures, section 6.9.
const (
	TPMSTRspCommand         TPMST = 0x00C4
	TPMSTNull               TPMST = 0x8000
	TPMSTNoSessions         TPMST = 0x8001
	TPMSTSessions           TPMST = 0x8002
	TPMSTAttestNV           TPMST = 0x8014
	TPMSTAttestCommandAudit TPMST = 0x8015
	TPMSTAttestSessionAudit TPMST = 0x8016
	TPMSTAttestCertify      TPMST = 0x8017
	TPMSTAttestQuote        TPMST = 0x8018
	TPMSTAttestTime         TPMST = 0x8019
	TPMSTAttestCreation     TPMST = 0x801A
	TPMSTAttestNVDigest     TPMST = 0x801C
	TPMSTCreation           TPMST = 0x8021
	TPMSTVerified           TPMST = 0x8022
	TPMSTAuthSecret         TPMST = 0x8023
	TPMSTHashCheck          TPMST = 0x8024
	TPMSTAuthSigned         TPMST = 0x8025
	TPMSTFuManifest         TPMST = 0x8029
)

// TPMSU represents a TPM_SU.
// See definition in Part 2: Structures, section 6.10.
type TPMSU uint16

// TPMSU values come from Part 2: Structures, section  6.10.
const (
	TPMSUClear TPMSU = 0x0000
	TPMSUState TPMSU = 0x0001
)

// TPMSE represents a TPM_SE.
// See definition in Part 2: Structures, section 6.11.
type TPMSE uint8

// TPMSE values come from Part 2: Structures, section 6.11.
const (
	TPMSEHMAC   TPMSE = 0x00
	TPMSEPolicy TPMSE = 0x01
	TPMSETrial  TPMSE = 0x03
)

// TPMCap represents a TPM_CAP.
// See definition in Part 2: Structures, section 6.12.
type TPMCap uint32

// TPMCap values come from Part 2: Structures, section 6.12.
const (
	TPMCapAlgs          TPMCap = 0x00000000
	TPMCapHandles       TPMCap = 0x00000001
	TPMCapCommands      TPMCap = 0x00000002
	TPMCapPPCommands    TPMCap = 0x00000003
	TPMCapAuditCommands TPMCap = 0x00000004
	TPMCapPCRs          TPMCap = 0x00000005
	TPMCapTPMProperties TPMCap = 0x00000006
	TPMCapPCRProperties TPMCap = 0x00000007
	TPMCapECCCurves     TPMCap = 0x00000008
	TPMCapAuthPolicies  TPMCap = 0x00000009
	TPMCapACT           TPMCap = 0x0000000A
)

// TPMPT represents a TPM_PT.
// See definition in Part 2: Structures, section 6.13.
type TPMPT uint32

// TPMPT values come from Part 2: Structures, section  6.13.
const (
	// a 4-octet character string containing the TPM Family value
	// (TPM_SPEC_FAMILY)
	TPMPTFamilyIndicator TPMPT = 0x00000100
	// the level of the specification
	TPMPTLevel TPMPT = 0x00000101
	// the specification Revision times 100
	TPMPTRevision TPMPT = 0x00000102
	// the specification day of year using TCG calendar
	TPMPTDayofYear TPMPT = 0x00000103
	// the specification year using the CE
	TPMPTYear TPMPT = 0x00000104
	// the vendor ID unique to each TPM manufacturer
	TPMPTManufacturer TPMPT = 0x00000105
	// the first four characters of the vendor ID string
	TPMPTVendorString1 TPMPT = 0x00000106
	// the second four characters of the vendor ID string
	TPMPTVendorString2 TPMPT = 0x00000107
	// the third four characters of the vendor ID string
	TPMPTVendorString3 TPMPT = 0x00000108
	// the fourth four characters of the vendor ID sting
	TPMPTVendorString4 TPMPT = 0x00000109
	// vendor-defined value indicating the TPM model
	TPMPTVendorTPMType TPMPT = 0x0000010A
	// the most-significant 32 bits of a TPM vendor-specific value
	// indicating the version number of the firmware.
	TPMPTFirmwareVersion1 TPMPT = 0x0000010B
	// the least-significant 32 bits of a TPM vendor-specific value
	// indicating the version number of the firmware.
	TPMPTFirmwareVersion2 TPMPT = 0x0000010C
	// the maximum size of a parameter TPM2B_MAX_BUFFER)
	TPMPTInputBuffer TPMPT = 0x0000010D
	// the minimum number of transient objects that can be held in TPM RAM
	TPMPTHRTransientMin TPMPT = 0x0000010E
	// the minimum number of persistent objects that can be held in TPM NV
	// memory
	TPMPTHRPersistentMin TPMPT = 0x0000010F
	// the minimum number of authorization sessions that can be held in TPM
	// RAM
	TPMPTHRLoadedMin TPMPT = 0x00000110
	// the number of authorization sessions that may be active at a time
	TPMPTActiveSessionsMax TPMPT = 0x00000111
	// the number of PCR implemented
	TPMPTPCRCount TPMPT = 0x00000112
	// the minimum number of octets in a TPMS_PCR_SELECT.sizeOfSelect
	TPMPTPCRSelectMin TPMPT = 0x00000113
	// the maximum allowed difference (unsigned) between the contextID
	// values of two saved session contexts
	TPMPTContextGapMax TPMPT = 0x00000114
	// the maximum number of NV Indexes that are allowed to have the
	// TPM_NT_COUNTER attribute
	TPMPTNVCountersMax TPMPT = 0x00000116
	// the maximum size of an NV Index data area
	TPMPTNVIndexMax TPMPT = 0x00000117
	// a TPMA_MEMORY indicating the memory management method for the TPM
	TPMPTMemory TPMPT = 0x00000118
	// interval, in milliseconds, between updates to the copy of
	// TPMS_CLOCK_INFO.clock in NV
	TPMPTClockUpdate TPMPT = 0x00000119
	// the algorithm used for the integrity HMAC on saved contexts and for
	// hashing the fuData of TPM2_FirmwareRead()
	TPMPTContextHash TPMPT = 0x0000011A
	// TPM_ALG_ID, the algorithm used for encryption of saved contexts
	TPMPTContextSym TPMPT = 0x0000011B
	// TPM_KEY_BITS, the size of the key used for encryption of saved
	// contexts
	TPMPTContextSymSize TPMPT = 0x0000011C
	// the modulus - 1 of the count for NV update of an orderly counter
	TPMPTOrderlyCount TPMPT = 0x0000011D
	// the maximum value for commandSize in a command
	TPMPTMaxCommandSize TPMPT = 0x0000011E
	// the maximum value for responseSize in a response
	TPMPTMaxResponseSize TPMPT = 0x0000011F
	// the maximum size of a digest that can be produced by the TPM
	TPMPTMaxDigest TPMPT = 0x00000120
	// the maximum size of an object context that will be returned by
	// TPM2_ContextSave
	TPMPTMaxObjectContext TPMPT = 0x00000121
	// the maximum size of a session context that will be returned by
	// TPM2_ContextSave
	TPMPTMaxSessionContext TPMPT = 0x00000122
	// platform-specific family (a TPM_PS value)(see Table 25)
	TPMPTPSFamilyIndicator TPMPT = 0x00000123
	// the level of the platform-specific specification
	TPMPTPSLevel TPMPT = 0x00000124
	// a platform specific value
	TPMPTPSRevision TPMPT = 0x00000125
	// the platform-specific TPM specification day of year using TCG
	// calendar
	TPMPTPSDayOfYear TPMPT = 0x00000126
	// the platform-specific TPM specification year using the CE
	TPMPTPSYear TPMPT = 0x00000127
	// the number of split signing operations supported by the TPM
	TPMPTSplitMax TPMPT = 0x00000128
	// total number of commands implemented in the TPM
	TPMPTTotalCommands TPMPT = 0x00000129
	// number of commands from the TPM library that are implemented
	TPMPTLibraryCommands TPMPT = 0x0000012A
	// number of vendor commands that are implemented
	TPMPTVendorCommands TPMPT = 0x0000012B
	// the maximum data size in one NV write, NV read, NV extend, or NV
	// certify command
	TPMPTNVBufferMax TPMPT = 0x0000012C
	// a TPMA_MODES value, indicating that the TPM is designed for these
	// modes.
	TPMPTModes TPMPT = 0x0000012D
	// the maximum size of a TPMS_CAPABILITY_DATA structure returned in
	// TPM2_GetCapability().
	TPMPTMaxCapBuffer TPMPT = 0x0000012E
	// TPMA_PERMANENT
	TPMPTPermanent TPMPT = 0x00000200
	// TPMA_STARTUP_CLEAR
	TPMPTStartupClear TPMPT = 0x00000201
	// the number of NV Indexes currently defined
	TPMPTHRNVIndex TPMPT = 0x00000202
	// the number of authorization sessions currently loaded into TPM RAM
	TPMPTHRLoaded TPMPT = 0x00000203
	// the number of additional authorization sessions, of any type, that
	// could be loaded into TPM RAM
	TPMPTHRLoadedAvail TPMPT = 0x00000204
	// the number of active authorization sessions currently being tracked
	// by the TPM
	TPMPTHRActive TPMPT = 0x00000205
	// the number of additional authorization sessions, of any type, that
	// could be created
	TPMPTHRActiveAvail TPMPT = 0x00000206
	// estimate of the number of additional transient objects that could be
	// loaded into TPM RAM
	TPMPTHRTransientAvail TPMPT = 0x00000207
	// the number of persistent objects currently loaded into TPM NV memory
	TPMPTHRPersistent TPMPT = 0x00000208
	// the number of additional persistent objects that could be loaded into
	// NV memory
	TPMPTHRPersistentAvail TPMPT = 0x00000209
	// the number of defined NV Indexes that have NV the TPM_NT_COUNTER
	// attribute
	TPMPTNVCounters TPMPT = 0x0000020A
	// the number of additional NV Indexes that can be defined with their
	// TPM_NT of TPM_NV_COUNTER and the TPMA_NV_ORDERLY attribute SET
	TPMPTNVCountersAvail TPMPT = 0x0000020B
	// code that limits the algorithms that may be used with the TPM
	TPMPTAlgorithmSet TPMPT = 0x0000020C
	// the number of loaded ECC curves
	TPMPTLoadedCurves TPMPT = 0x0000020D
	// the current value of the lockout counter (failedTries)
	TPMPTLockoutCounter TPMPT = 0x0000020E
	// the number of authorization failures before DA lockout is invoked
	TPMPTMaxAuthFail TPMPT = 0x0000020F
	// the number of seconds before the value reported by
	// TPM_PT_LOCKOUT_COUNTER is decremented
	TPMPTLockoutInterval TPMPT = 0x00000210
	// the number of seconds after a lockoutAuth failure before use of
	// lockoutAuth may be attempted again
	TPMPTLockoutRecovery TPMPT = 0x00000211
	// number of milliseconds before the TPM will accept another command
	// that will modify NV
	TPMPTNVWriteRecovery TPMPT = 0x00000212
	// the high-order 32 bits of the command audit counter
	TPMPTAuditCounter0 TPMPT = 0x00000213
	// the low-order 32 bits of the command audit counter
	TPMPTAuditCounter1 TPMPT = 0x00000214
)

// TPMPTPCR represents a TPM_PT_PCR.
// See definition in Part 2: Structures, section 6.14.
type TPMPTPCR uint32

// TPMPTPCR values come from Part 2: Structures, section 6.14.
const (
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR is saved and
	// restored by TPM_SU_STATE
	TPMPTPCRSave TPMPTPCR = 0x00000000
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be
	// extended from locality 0
	TPMPTPCRExtendL0 TPMPTPCR = 0x00000001
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be reset
	// by TPM2_PCR_Reset() from locality 0
	TPMPTPCRResetL0 TPMPTPCR = 0x00000002
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be
	// extended from locality 1
	TPMPTPCRExtendL1 TPMPTPCR = 0x00000003
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be reset
	// by TPM2_PCR_Reset() from locality 1
	TPMPTPCRResetL1 TPMPTPCR = 0x00000004
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be
	// extended from locality 2
	TPMPTPCRExtendL2 TPMPTPCR = 0x00000005
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be reset
	// by TPM2_PCR_Reset() from locality 2
	TPMPTPCRResetL2 TPMPTPCR = 0x00000006
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be
	// extended from locality 3
	TPMPTPCRExtendL3 TPMPTPCR = 0x00000007
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be reset
	// by TPM2_PCR_Reset() from locality 3
	TPMPTPCRResetL3 TPMPTPCR = 0x00000008
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be
	// extended from locality 4
	TPMPTPCRExtendL4 TPMPTPCR = 0x00000009
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR may be reset
	// by TPM2_PCR_Reset() from locality 4
	TPMPTPCRResetL4 TPMPTPCR = 0x0000000A
	// a SET bit in the TPMS_PCR_SELECT indicates that modifications to this
	// PCR (reset or Extend) will not increment the pcrUpdateCounter
	TPMPTPCRNoIncrement TPMPTPCR = 0x00000011
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR is reset by a
	// D-RTM event
	TPMPTPCRDRTMRest TPMPTPCR = 0x00000012
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR is controlled
	// by policy
	TPMPTPCRPolicy TPMPTPCR = 0x00000013
	// a SET bit in the TPMS_PCR_SELECT indicates that the PCR is controlled
	// by an authorization value
	TPMPTPCRAuth TPMPTPCR = 0x00000014
)

// TPMHT represents a TPM_HT.
// See definition in Part 2: Structures, section 7.2.
type TPMHT uint8

// TPMHT values come from Part 2: Structures, section 7.2.
const (
	TPMHTPCR           TPMHT = 0x00
	TPMHTNVIndex       TPMHT = 0x01
	TPMHTHMACSession   TPMHT = 0x02
	TPMHTPolicySession TPMHT = 0x03
	TPMHTPermanent     TPMHT = 0x40
	TPMHTTransient     TPMHT = 0x80
	TPMHTPersistent    TPMHT = 0x81
	TPMHTAC            TPMHT = 0x90
)

// Saved Context transient object handles.
// See definition in Part 2: Structures, section 14.6.2
// Context Handle Values come from table 211
const (
	// an ordinary transient object
	TPMIDHSavedTransient TPMIDHSaved = 0x80000000
	// a sequence object
	TPMIDHSavedSequence TPMIDHSaved = 0x80000001
	// a transient object with the stClear attribute SET
	TPMIDHSavedTransientClear TPMIDHSaved = 0x80000002
)

// TPMHandle represents a TPM_HANDLE.
// See definition in Part 2: Structures, section 7.1.
type TPMHandle uint32

// TPMHandle values come from Part 2: Structures, section 7.4.
const (
	TPMRHOwner         TPMHandle = 0x40000001
	TPMRHNull          TPMHandle = 0x40000007
	TPMRSPW            TPMHandle = 0x40000009
	TPMRHLockout       TPMHandle = 0x4000000A
	TPMRHEndorsement   TPMHandle = 0x4000000B
	TPMRHPlatform      TPMHandle = 0x4000000C
	TPMRHPlatformNV    TPMHandle = 0x4000000D
	TPMRHFWOwner       TPMHandle = 0x40000140
	TPMRHFWEndorsement TPMHandle = 0x40000141
	TPMRHFWPlatform    TPMHandle = 0x40000142
	TPMRHFWNull        TPMHandle = 0x40000143
)

// TPMNT represents a TPM_NT.
// See definition in Part 2: Structures, section 13.4.
type TPMNT uint8

// TPMNT values come from Part 2: Structures, section 13.2.
const (
	// contains data that is opaque to the TPM that can only be modified
	// using TPM2_NV_Write().
	TPMNTOrdinary TPMNT = 0x0
	// contains an 8-octet value that is to be used as a counter and can
	// only be modified with TPM2_NV_Increment()
	TPMNTCounter TPMNT = 0x1
	// contains an 8-octet value to be used as a bit field and can only be
	// modified with TPM2_NV_SetBits().
	TPMNTBits TPMNT = 0x2
	// contains a digest-sized value used like a PCR. The Index can only be
	// modified using TPM2_NV_Extend(). The extend will use the nameAlg of
	// the Index.
	TPMNTExtend TPMNT = 0x4
	// contains pinCount that increments on a PIN authorization failure and
	// a pinLimit
	TPMNTPinFail TPMNT = 0x8
	// contains pinCount that increments on a PIN authorization success and
	// a pinLimit
	TPMNTPinPass TPMNT = 0x9
)
