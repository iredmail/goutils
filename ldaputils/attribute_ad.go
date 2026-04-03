package ldaputils

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

const (
	// AD attributes

	AttrObjectGUID  = "objectGUID"
	AttrObjectSid   = "objectSid"
	AttrWhenCreated = "whenCreated"
	AttrWhenChanged = "whenChanged"
)

// EpochToADTimestamp 将 epoch 转换为 AD 使用的时间戳格式。
func EpochToADTimestamp(epoch int64) string {
	t := time.Unix(epoch, 0)

	return t.Truncate(time.Second).UTC().Format("20060102150405.0Z")
}

func GetAttrObjectGUID(entry *ldap.Entry) (guid string) {
	guidBytes := entry.GetRawAttributeValue(AttrObjectGUID)

	if len(guidBytes) != 16 {
		return
	}

	return fmt.Sprintf(
		"%08x-%04x-%04x-%02x%02x-%x",
		binary.LittleEndian.Uint32(guidBytes[0:4]),
		binary.LittleEndian.Uint16(guidBytes[4:6]),
		binary.LittleEndian.Uint16(guidBytes[6:8]),
		guidBytes[8], guidBytes[9], guidBytes[10:16],
	)
}

// GetAttrObjectSID 从 LDAP 条目中获取 objectSid 属性的值，并将其转换为字符串格式。
func GetAttrObjectSID(entry *ldap.Entry) string {
	sidBytes := entry.GetRawAttributeValue(AttrObjectSid)

	if len(sidBytes) < 8 {
		return ""
	}

	revision := int(sidBytes[0])
	subAuthorityCount := int(sidBytes[1])

	// 标识符授权机构
	identifierAuthority := uint64(0)
	for i := 2; i < 8; i++ {
		identifierAuthority = (identifierAuthority << 8) | uint64(sidBytes[i])
	}

	var sidBuilder strings.Builder
	sidBuilder.WriteString("S-")
	sidBuilder.WriteString(strconv.Itoa(revision))
	sidBuilder.WriteString("-")
	sidBuilder.WriteString(strconv.FormatUint(identifierAuthority, 10))

	// 子授权机构
	if len(sidBytes) < 8+4*subAuthorityCount {
		return ""
	}

	for i := 0; i < subAuthorityCount; i++ {
		start := 8 + i*4
		subAuthority := binary.LittleEndian.Uint32(sidBytes[start : start+4])
		sidBuilder.WriteString("-")
		sidBuilder.WriteString(strconv.FormatUint(uint64(subAuthority), 10))
	}

	return sidBuilder.String()
}
