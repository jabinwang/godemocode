package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Config struct {
	Name string `json:"server-name"`
	IP string `json:"server-ip"`
}

func readConfig() *Config {
	config:= Config{}
	typ := reflect.TypeOf(config)
	fmt.Printf("type %v, %v\n", typ, typ.Kind())
	value := reflect.Indirect(reflect.ValueOf(&config))
	fmt.Printf("value %v\n", value)
	for i := 0; i< typ.NumField(); i++ {
		f := typ.Field(i)
		fmt.Printf("field %v\n", f)
		if v, ok := f.Tag.Lookup("json"); ok {
			key := fmt.Sprintf("CONFIG_%s", strings.ReplaceAll(strings.ToUpper(v), "-", "_"))
			if env, exist := os.LookupEnv(key); exist {
				fmt.Printf("env %v, field name %v\n", env, value.FieldByName(f.Name).String())
				value.FieldByName(f.Name).Set(reflect.ValueOf(env))
			}
		}
	}
	return &config
}
func main()  {
	os.Setenv("CONFIG_SERVER_NAME", "global_server")
	os.Setenv("CONFIG_SERVER_IP", "10.0.0.1")
	c := readConfig()
	fmt.Printf("%v\n", c)
}
