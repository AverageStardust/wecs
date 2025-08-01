package common

import "hash/crc64"

var Crc64ISOTable = crc64.MakeTable(crc64.ISO)
