package syntropy

import (
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"time"
)

func NullableStringToString(s syntropy.NullableString) string {
	val := s.Get()
	if val == nil {
		return ""
	}
	return *val
}

func NullableAgentStatusToString(s syntropy.NullableAgentStatus) string {
	val := s.Get()
	if val == nil {
		return ""
	}
	return string(*val)
}

func convertAgentTagsToTfValue(in []syntropy.AgentTag) []Tag {
	var out []Tag
	for _, tag := range in {
		out = append(out, Tag{
			ID:   int64(tag.AgentTagId),
			Name: tag.AgentTagName,
		})
	}
	return out
}

func int64ArrayToInt32Array(arr []int64) []int32 {
	ret := make([]int32, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, int32(v))
	}
	return ret
}

func stringArrayToAgentTypeArray(arr []string) []syntropy.AgentType {
	ret := make([]syntropy.AgentType, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, syntropy.AgentType(v))
	}
	return ret
}

func stringArrayToAgentStatusArray(arr []string) []syntropy.AgentFilterAgentStatus {
	ret := make([]syntropy.AgentFilterAgentStatus, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, syntropy.AgentFilterAgentStatus(v))
	}
	return ret
}

func tfValueToDateP(date string) (*time.Time, error) {
	layout := "2014-09-12T11:45:26.371Z"
	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func sumOfNaturalNumbers(n int) (sum int) {
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}
