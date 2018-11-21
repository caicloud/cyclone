/*
Copyright 2017 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validator

// Tag is the validation tag can be used.
type Tag string

// Tags for ref and doc gen.
const (
	TagIsColor               Tag = "iscolor"
	TagHasValue              Tag = "required"
	TagIsDefault             Tag = "isdefault"
	TagHasLengthOf           Tag = "len"
	TagHasMinOf              Tag = "min"
	TagHasMaxOf              Tag = "max"
	TagIsEq                  Tag = "eq"
	TagIsNe                  Tag = "ne"
	TagIsLt                  Tag = "lt"
	TagIsLte                 Tag = "lte"
	TagIsGt                  Tag = "gt"
	TagIsGte                 Tag = "gte"
	TagIsEqField             Tag = "eqfield"
	TagIsEqCrossStructField  Tag = "eqcsfield"
	TagIsNeCrossStructField  Tag = "necsfield"
	TagIsGtCrossStructField  Tag = "gtcsfield"
	TagIsGteCrossStructField Tag = "gtecsfield"
	TagIsLtCrossStructField  Tag = "ltcsfield"
	TagIsLteCrossStructField Tag = "ltecsfield"
	TagIsNeField             Tag = "nefield"
	TagIsGteField            Tag = "gtefield"
	TagIsGtField             Tag = "gtfield"
	TagIsLteField            Tag = "ltefield"
	TagIsLtField             Tag = "ltfield"
	TagIsAlpha               Tag = "alpha"
	TagIsAlphanum            Tag = "alphanum"
	TagIsAlphaUnicode        Tag = "alphaunicode"
	TagIsAlphanumUnicode     Tag = "alphanumunicode"
	TagIsNumeric             Tag = "numeric"
	TagIsNumber              Tag = "number"
	TagIsHexadecimal         Tag = "hexadecimal"
	TagIsHEXColor            Tag = "hexcolor"
	TagIsRGB                 Tag = "rgb"
	TagIsRGBA                Tag = "rgba"
	TagIsHSL                 Tag = "hsl"
	TagIsHSLA                Tag = "hsla"
	TagIsEmail               Tag = "email"
	TagIsURL                 Tag = "url"
	TagIsURI                 Tag = "uri"
	TagIsBase64              Tag = "base64"
	TagContains              Tag = "contains"
	TagContainsAny           Tag = "containsany"
	TagContainsRune          Tag = "containsrune"
	TagExcludes              Tag = "excludes"
	TagExcludesAll           Tag = "excludesall"
	TagExcludesRune          Tag = "excludesrune"
	TagIsISBN                Tag = "isbn"
	TagIsISBN10              Tag = "isbn10"
	TagIsISBN13              Tag = "isbn13"
	TagIsUUID                Tag = "uuid"
	TagIsUUID3               Tag = "uuid3"
	TagIsUUID4               Tag = "uuid4"
	TagIsUUID5               Tag = "uuid5"
	TagIsASCII               Tag = "ascii"
	TagIsPrintableASCII      Tag = "printascii"
	TagHasMultiByteCharacter Tag = "multibyte"
	TagIsDataURI             Tag = "datauri"
	TagIsLatitude            Tag = "latitude"
	TagIsLongitude           Tag = "longitude"
	TagIsSSN                 Tag = "ssn"
	TagIsIPv4                Tag = "ipv4"
	TagIsIPv6                Tag = "ipv6"
	TagIsIP                  Tag = "ip"
	TagIsCIDRv4              Tag = "cidrv4"
	TagIsCIDRv6              Tag = "cidrv6"
	TagIsCIDR                Tag = "cidr"
	TagIsTCP4AddrResolvable  Tag = "tcp4_addr"
	TagIsTCP6AddrResolvable  Tag = "tcp6_addr"
	TagIsTCPAddrResolvable   Tag = "tcp_addr"
	TagIsUDP4AddrResolvable  Tag = "udp4_addr"
	TagIsUDP6AddrResolvable  Tag = "udp6_addr"
	TagIsUDPAddrResolvable   Tag = "udp_addr"
	TagIsIP4AddrResolvable   Tag = "ip4_addr"
	TagIsIP6AddrResolvable   Tag = "ip6_addr"
	TagIsIPAddrResolvable    Tag = "ip_addr"
	TagIsUnixAddrResolvable  Tag = "unix_addr"
	TagIsMAC                 Tag = "mac"
	TagIsHostname            Tag = "hostname"
	TagIsFQDN                Tag = "fqdn"
	TagIsUnique              Tag = "unique"
)

// Special tags.
const (
	TagUTF8HexComma       Tag = "0x2C"
	TagUTF8Pipe           Tag = "0x7C"
	TagAndSeparator       Tag = ","
	TagOrSeparator        Tag = "|"
	TagKeySeparator       Tag = "="
	TagStructOnly         Tag = "structonly"
	TagNoStructLevel      Tag = "nostructlevel"
	TagOmitempty          Tag = "omitempty"
	TagSkipValidationTag  Tag = "-"
	TagDiveTag            Tag = "dive"
	TagKeysTag            Tag = "keys"
	TagEndKeysTag         Tag = "endkeys"
	TagRequiredTag        Tag = "required"
	TagNamespaceSeparator Tag = "."
	TagLeftBracket        Tag = "["
	TagRightBracket       Tag = "]"
)
