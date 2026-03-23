package service

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	randomMu     sync.Mutex
)

const randomCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var nickAdjectives = []string{
	"눈부신", "따뜻한", "우수한", "은밀한", "침착한",
	"잠든", "풍부한", "환상적인", "고요한", "느긋한",
	"독특한", "위대한", "미묘한", "섬세한", "즐거운",
	"행복한", "고독한", "신비로운", "찬란한", "조용한",
	"빛나는", "화려한", "평화로운", "우아한", "뜨거운",
	"차가운", "부드러운", "귀여운", "발랄한", "활발한",
}

var nickAnimals = []string{
	"얼룩말", "양", "낙타", "사막여우", "기린",
	"코끼리", "하마", "코알라", "나무늘보", "호랑이",
	"사자", "부엉이", "고래", "상어", "개구리",
	"구피", "고양이", "강아지", "햄스터", "카피바라",
	"쿼카", "판다", "거북이", "토끼", "불가사리",
	"해파리", "미어캣", "도마뱀", "기니피그", "사슴",
}

func generateRandomString(length int) string {
	randomMu.Lock()
	defer randomMu.Unlock()

	result := make([]byte, length)
	for i := range result {
		result[i] = randomCharacters[randomSource.Intn(len(randomCharacters))]
	}

	return string(result)
}

func generateRandomNick() string {
	randomMu.Lock()
	defer randomMu.Unlock()

	return nickAdjectives[randomSource.Intn(len(nickAdjectives))] + nickAnimals[randomSource.Intn(len(nickAnimals))]
}
