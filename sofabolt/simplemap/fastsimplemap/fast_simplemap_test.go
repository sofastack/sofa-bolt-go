// nolint
// Copyright 20xx The Alipay Authors.
//
// @authors[0]: bingwu.ybw(bingwu.ybw@antfin.com|detailyang@gmail.com)
// @authors[1]: robotx(robotx@antfin.com)
//
// *Legal Disclaimer*
// Within this source code, the comments in Chinese shall be the original, governing version. Any comment in other languages are for reference only. In the event of any conflict between the Chinese language version comments and other language version comments, the Chinese language version shall prevail.
// *法律免责声明*
// 关于代码注释部分，中文注释为官方版本，其它语言注释仅做参考。中文注释可能与其它语言注释存在不一致，当中文注释与其它语言注释存在不一致时，请以中文注释为准。
//
//

package fastsimplemap

import (
	"fmt"
	"testing"
)

func TestFastSimpleMap(t *testing.T) {
	var sm FastSimpleMap

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		sm.Set(ii, ii)
	}

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		if sm.Get(ii) != ii {
			t.Fatal("expect equal")
		}
	}

	sm.Reset()

	for i := 0; i < 1024; i++ {
		ii := fmt.Sprintf("%d", i)
		sm.Set(ii, ii)
	}

	for i := 0; i < 1024; i++ {
		ii := fmt.Sprintf("%d", i)
		if sm.Get(ii) != ii {
			t.Fatal("expect equal")
		}
	}

	sm.Set("a", "abcd")
	sm.Set("a", "defg")
	if sm.Get("a") != "defg" {
		t.Fatal("expcet defg")
	}
}

func TestFastSimpleMapDel(t *testing.T) {
	var sm FastSimpleMap

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		sm.Set(ii, ii)
	}

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		sm.Del(ii)
		if sm.Get(ii) != "" {
			t.Fatal("expect empty but got", sm.Get(ii))
		}
	}
}

func TestFastSimpleMapCopy(t *testing.T) {
	var sm FastSimpleMap

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		sm.Set(ii, ii)
	}

	var bm FastSimpleMap

	sm.CopyTo(&bm)

	for i := 0; i < 100; i++ {
		ii := fmt.Sprintf("%d", i)
		if bm.Get(ii) != ii {
			t.Fatal("expect equal")
		}
	}
}
