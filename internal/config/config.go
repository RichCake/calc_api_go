package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr string
	TimeAdd time.Duration
	TimeSub time.Duration
	TimeMul time.Duration
	TimeDiv time.Duration
	Power int
}

func ConfigFromEnv() *Config {
	config := new(Config)

	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}

	duration_string := os.Getenv("TIME_ADDITION_MS")
	duration, err := time.ParseDuration(duration_string + "ms")
	if duration_string == "" || err != nil {
		config.TimeAdd = time.Second
	} else {
		config.TimeAdd = duration
	}

	duration_string = os.Getenv("TIME_SUBTRACTION_MS")
	duration, err = time.ParseDuration(duration_string + "ms")
	if duration_string == "" || err != nil {
		config.TimeSub = time.Second
	} else {
		config.TimeSub = duration
	}

	duration_string = os.Getenv("TIME_MULTIPLICATIONS_MS")
	duration, err = time.ParseDuration(duration_string + "ms")
	if duration_string == "" || err != nil {
		config.TimeMul = time.Second
	} else {
		config.TimeMul = duration
	}

	duration_string = os.Getenv("TIME_DIVISIONS_MS")
	duration, err = time.ParseDuration(duration_string + "ms")
	if duration_string == "" || err != nil {
		config.TimeDiv = time.Second
	} else {
		config.TimeDiv = duration
	}

	power_str := os.Getenv("COMPUTING_POWER")
	power_int, err := strconv.Atoi(power_str)
	if power_str == "" || err != nil{
		config.Power = 5
	} else {
		config.Power = power_int
	}

	return config
}