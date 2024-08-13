package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
)

var pinyinArgs pinyin.Args

func init() {
	pinyinArgs = pinyin.NewArgs()
	pinyinArgs.Style = pinyin.FirstLetter
	pinyinArgs.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}
}

// Sha256Hash 计算SHA 256 小写哈兮
func Sha256Hash(txt string) string {
	m := sha256.New()
	m.Write([]byte(txt))
	return hex.EncodeToString(m.Sum(nil))
}

// MacOk MAC是否正确
func MacOk(txt string) bool {
	regular := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})|([0-9a-fA-F]{4}\\.[0-9a-fA-F]{4}\\.[0-9a-fA-F]{4})$`
	reg := regexp.MustCompile(regular)
	return reg.MatchString(txt)
}

// IpOk IP是否正确
func IpOk(txt string) bool {
	regular := `^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	reg := regexp.MustCompile(regular)
	return reg.MatchString(txt)
}

// UrlOk 网址是否正确
func UrlOk(txt string) bool {
	re := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|\/|\/\/)?[A-z0-9_-]*?[:]?[A-z0-9_-]*?[@]?[A-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
	return re.MatchString(txt)
}

// EmailOk 电子邮箱是否正确
func EmailOk(txt string) bool {
	_, e := mail.ParseAddress(txt)
	return e == nil
}

// MobilePhoneOk 手机号码是否正确
func MobilePhoneOk(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147)|(888)|(58[0-9]))\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

// ArrayIndex 返回值在数组中的索引
//
// 没有时索引为-1
func ArrayIndex(v string, a []string) int {
	for ai, av := range a {
		if v == av {
			return ai
		}
	}

	return -1
}

// ToPinyinWithFirstLetter 转为拼音首字母(非汉字全文保留)
func ToPinyinWithFirstLetter(txt string) string {
	return strings.Join(pinyin.LazyPinyin(txt, pinyinArgs), "")
}

// 文本转时间(例如 2015-02-25 11:06:39)转时间
//
// 转换失败时为nil
//
// 提示: Linux下time.Parse()得到的是UTC时区的时间, 所以此处用time.ParseInLocation()
func DtTextToTime(dt string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
}

// 是否不是日期时间文本
func IsNotDateTimeText(dt string) bool {
	if dt == "" {
		return true
	}

	_, e := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
	return e != nil
}
