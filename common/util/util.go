package util

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strconv"
)

func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()

	v.SetConfigFile("json")
	v.AddConfigPath(path)
	v.SetConfigName(filename)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("unmarshal json fail, err:%v", err)
		return err
	}

	return nil
}

func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("unmarshal env fail, err:%v", err)
		return err
	}

	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool())
		default:
			panic("Unsupported type")
		}

		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("set env fail, err:%v", err)
			return err
		}
	}

	return nil
}

func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()
	v.SetConfigType("json")
	err := v.AddRemoteProvider("consol", endPoint, path)
	if err != nil {
		logrus.Errorf("add remote provider fail, err:%v", err)
		return err
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("read remote config fail, err:%v", err)
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("unmarshal json fail, err:%v", err)
		return err
	}

	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("set env fail, err:%v", err)
		return err
	}

	return nil
}
