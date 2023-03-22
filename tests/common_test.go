package tests

import (
	"fmt"
	"reviewmakerback/common"
	"testing"
)

func TestSubstring(t *testing.T) {
	if common.Substring("123456789", 0, 1) != "1" {
		t.Error("miss")
	}
	if common.Substring("123456789", 0, 3) != "123" {
		t.Error("miss")
	}
	if common.Substring("123456789", 1, 1) != "2" {
		t.Error("miss")
	}
	if common.Substring("123456789", 1, 5) != "23456" {
		t.Error("miss")
	}
	if common.Substring("123456789", 0, 9) != "123456789" {
		t.Error("miss")
	}
	if common.Substring("123456789", 0, 10) != "123456789" {
		t.Error("miss")
	}
}

func TestSubstringMult(t *testing.T) {
	if s := common.SubstringMult("１２３４５６７８９", 0, 1); s != "１" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１２３４５６７８９", 0, 3); s != "１２３" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１２３４５６７８９", 1, 1); s != "２" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１２３４５６７８９", 1, 5); s != "２３４５６" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１２３４５６７８９", 0, 9); s != "１２３４５６７８９" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１２３４５６７８９", 0, 10); s != "１２３４５６７８９" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
	if s := common.SubstringMult("１2３4５5７8９", 0, 3); s != "１2３" {
		t.Error(fmt.Sprintf("miss '%s'", s))
	}
}

func TestCommonSplitQueryChain1(t *testing.T) {
	m := common.SplitQueryChain("abcd=123&efgh=456")

	s1 := m["abcd"]
	s2 := m["efgh"]

	if s1 != "123" {
		t.Error("miss")
	}
	if s2 != "456" {
		t.Error("miss")
	}
}

func TestCommonSplitQueryChain2(t *testing.T) {
	m := common.SplitQueryChain("abcd=&efgh=")

	s1 := m["abcd"]
	s2 := m["efgh"]

	if s1 != "" {
		t.Error("miss")
	}
	if s2 != "" {
		t.Error("miss")
	}
}

func TestCommonSplitQueryChain3(t *testing.T) {
	m := common.SplitQueryChain("=123&")

	_, ok1 := m["abcd"]
	_, ok2 := m["efgh"]

	if ok1 == true {
		t.Error("miss")
	}
	if ok2 == true {
		t.Error("miss")
	}
}

func TestCommonRegexp(t *testing.T) {
	if !common.TestRegexp(`^[a-zA-Z0-9._]*$`, "abc123ABC") {
		t.Error("miss")
	}
	if common.TestRegexp(`^[a-zA-Z0-9._]*$`, "abc123!ABC") {
		t.Error("miss")
	}
	if common.TestRegexp(`^[a-zA-Z0-9._]*$`, "abc/123/ABC") {
		t.Error("miss")
	}
	if common.TestRegexp(`^[a-zA-Z0-9._]*$`, "*") {
		t.Error("miss")
	}
	if common.TestRegexp(`^[a-zA-Z0-9._]*$`, "../aaa") {
		t.Error("miss")
	}
	if common.TestRegexp(`^[a-zA-Z0-9._]*$`, "..\\aaa") {
		t.Error("miss")
	}
}
