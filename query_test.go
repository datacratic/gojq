// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"encoding/json"
	"log"
	"testing"
)

var sample = []byte(`{"bid":{"request":{"at":2,"bcat":["IAB7-41","IAB11-4","IAB8-5","IAB14","IAB7","IAB14-1","IAB25-2","IAB24","IAB13-7","IAB16"],"cur":["USD"],"device":{"ip":"208.88.255.234","language":"EN","ua":"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36"},"ext":{"edepth":1,"exchange":"datacratic","sdepth":1,"ssl":0},"id":"D2E99F80C48439FE","imp":[{"banner":{"h":250,"pos":1,"topframe":1,"w":300},"ext":{"creative-ids":{"4":[1],"5":[1]}},"id":"1","secure":0}],"site":{"cat":["IAB12"],"content":{"keywords":"2"},"id":"168042","page":"http://www.nytimes.com/2015/12/03/fashion/the-2016-pirelli-calendar-may-signal-a-cultural-shift.html?smid=fb-nytimes&smtyp=cur&_r=0","publisher":{"id":"183760"},"ref":"https://www.facebook.com/"},"tmax":79,"user":{"id":"VFz6X8AoI3oAAEViQ-MAAAH8"}},"response":{"seatbid":[{"bid":[{"id":"D2E99F80C48439FE:1","impid":"1","price":1500,"cid":"5","crid":"1","ext":{"bidder":{"campaign":"datacratic_01","strategy":"datacratic_01_01"}}}]}]}},"win":{"impid":"1","price":1000}}`)

func TestExtractAndFindJSON(t *testing.T) {
	v := Value{}
	if err := v.Unmarshal(sample); err != nil {
		t.Fatal(err)
		return
	}

	q := v.NewQuery()
	ref, err := q.String("bid", "request", "site", "ref")
	if err != nil {
		t.Fatal(err)
	}

	if ref != "https://www.facebook.com/" {
		t.Fail()
	}

	err = q.Array("bid", "request", "imp")
	if err != nil {
		t.Fatal(err)
	}

	err = q.ForEach(func(i Query) error {
		w, err := i.Int64("banner", "w")
		if err != nil {
			t.Fatal(err)
		}

		log.Println(w)
		if w != 300 {
			t.Fail()
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkExtractGoAndFindJSON(b *testing.B) {
	var m interface{}
	if err := json.Unmarshal(sample, &m); err != nil {
		log.Println(err)
		return
	}

	b.ResetTimer()
	count := 0

	for i := 0; i < b.N; i++ {
		va, ok := m.(map[string]interface{})
		if !ok {
			b.Fail()
		}

		vb, ok := va["bid"]
		if !ok {
			b.Fail()
		}

		vc, ok := vb.(map[string]interface{})
		if !ok {
			b.Fail()
		}

		vd, ok := vc["request"]
		if !ok {
			b.Fail()
		}

		ve, ok := vd.(map[string]interface{})
		if !ok {
			b.Fail()
		}

		vf, ok := ve["site"]
		if !ok {
			b.Fail()
		}

		vg, ok := vf.(map[string]interface{})
		if !ok {
			b.Fail()
		}

		vh, ok := vg["ref"]
		if !ok {
			b.Fail()
		}

		vi, ok := vh.(string)
		if !ok {
			b.Fail()
		}

		if vi != "" {
			count += len(vi)
		}
	}

	if count == 0 {
		b.Fail()
	}
}

func BenchmarkExtractAndFindJSONFast(b *testing.B) {
	v := Value{}
	if err := v.Unmarshal(sample); err != nil {
		log.Println(err)
		return
	}

	b.ResetTimer()
	count := 0

	for i := 0; i < b.N; i++ {
		q := v.NewQuery()
		domain, err := q.String("bid", "request", "site", "ref")
		if err != nil {
			return
		}

		if domain != "" {
			count += len(domain)
		}
	}

	if count == 0 {
		b.Fail()
	}
}
