package xmlio

import (
	"io"
	"strings"
)

func WriteCDATA(w io.Writer, data []byte) error {
	if _, err := w.Write([]byte("<![CDATA[")); err != nil {
		return err
	}
	i := 0
	for i < len(data) {
		if i+2 < len(data) && data[i] == ']' && data[i+1] == ']' && data[i+2] == '>' {
			if _, err := w.Write([]byte("]]]]><![CDATA[>")); err != nil {
				return err
			}
			i += 3
		} else {
			if _, err := w.Write(data[i : i+1]); err != nil {
				return err
			}
			i++
		}
	}
	_, err := w.Write([]byte("]]>"))
	return err
}

func EscapeAttr(val string) string {
	res := strings.ReplaceAll(val, "&", "&amp;")
	res = strings.ReplaceAll(res, "\"", "&quot;")
	res = strings.ReplaceAll(res, "<", "&lt;")
	res = strings.ReplaceAll(res, ">", "&gt;")
	return res
}

func UnescapeAttr(val string) string {
	res := strings.ReplaceAll(val, "&quot;", "\"")
	res = strings.ReplaceAll(res, "&amp;", "&")
	res = strings.ReplaceAll(res, "&lt;", "<")
	res = strings.ReplaceAll(res, "&gt;", ">")
	return res
}
