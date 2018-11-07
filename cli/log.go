package cli

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Log(a ...interface{}) {
	if !viper.GetBool("quiet") {
		fmt.Println(a...)
	}
}

func Logf(format string, a ...interface{}) {
	if !viper.GetBool("quiet") {
		fmt.Printf(format, a...)
	}
}

func Logerr(a ...interface{}) {
	if !viper.GetBool("quiet") {
		fmt.Fprintln(os.Stderr, a...)
	}
}

func Logerrf(format string, a ...interface{}) {
	if !viper.GetBool("quiet") {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}
