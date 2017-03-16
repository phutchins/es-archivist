package main

import "testing"

func TestGetIndexArray(t *testing.T) {

}

type testArray struct {
  values []string
  sortedValues []string
}

var tests = []testArray{
  { []string{"logstash-2016.12.16", "logstash-2016.11.01", "logstash-2017.01.03", },
    []string{"logstash-2016.11.01", "logstash-2016.12.16", "logstash-2017.01.03", } },
  { []string{"storj-2016.12.16", "logstash-2016.11.01", "logstash-2017.01.03", },
    []string{"logstash-2016.11.01", "storj-2016.12.16", "logstash-2017.01.03", } },
}

func TestSortIndexArrayLength(t *testing.T) {
  t.Log("Comparing sorted index array length")

  for _, indexArray := range tests {
    v := SortIndexArray(indexArray.values)

    arraysAreEqual, expected, got := CompareSliceLength(v, indexArray.sortedValues)

    if arraysAreEqual != true {
      t.Error(
        "For", indexArray.values,
        "expected", expected,
        "got", got,
      )
    }
  }
}

func TestSortIndexArrayContents(t *testing.T) {
  t.Log("Comparing sorted index content equality and order")

  for _, indexArray := range tests {
    v := SortIndexArray(indexArray.values)
    arraysSameContents, expected, got := CompareSliceContents(v, indexArray.sortedValues)

    if arraysSameContents != true {
      t.Error(
        "For", "array length comparison",
        "expected", expected,
        "got", got,
      )
    }
  }
}

func CompareSliceLength(X, Y []string) (bool, int, int) {
  xLen := len(X)
  yLen := len(Y)

  if ( xLen != yLen ) {
    return false, xLen, yLen
  }

  return true, 0, 0
}

func CompareSliceContents(X, Y []string) (bool, string, string) {
  for i, val := range X {
    if val != Y[i] {
      return false, val, Y[i]
    }
  }

  return true, "", ""
}
