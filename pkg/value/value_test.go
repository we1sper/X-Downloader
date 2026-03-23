package value

import (
	"fmt"
	"testing"
)

func prepare() *Value {
	data := map[string]interface{}{
		"user": "welsper",
		"emails": []string{
			"welsper@qq.com",
			"welsper@nit.com",
		},
		"details": map[string]interface{}{
			"age":     18,
			"country": "China",
			"job": map[string]string{
				"company": "China Merchants Bank",
				"salary":  "￥9,999,999/m",
			},
		},
	}

	return NewValue(data)
}

func TestValue_Trace(_ *testing.T) {
	value := prepare()

	fmt.Println(value.Get("user").Trace())
	fmt.Println(value.Get("details", "country").Trace())
	fmt.Println(value.Get("details", "job", "company").Trace())
	fmt.Println(value.Get("emails").ValueAt(1).Trace())
}

func TestValue_Debug(_ *testing.T) {
	value := prepare()

	fmt.Println(value.Get("user").Debug())
	fmt.Println(value.Get("details", "country").Debug())
	fmt.Println(value.Get("details", "job", "company").Debug())
	fmt.Println(value.Get("school").Debug())
	fmt.Println(value.Get("details", "location").Debug())
	fmt.Println(value.Get("details", "location", "province").Debug())
	fmt.Println(value.Get("emails").Head().Debug())
	fmt.Println(value.Get("emails").Tail().Debug())
	fmt.Println(value.Get("emails").ValueAt(0).Debug())
	fmt.Println(value.Get("emails").ValueAt(1).Debug())
	fmt.Println(value.Get("emails").ValueAt(2).Debug())
	fmt.Println(value.Get("emails").ValueAt(3).Debug())
	fmt.Println(value.Get("emails").ValueAt(0, 1).Debug())
}
