package main
import (
        "github.com/cihub/seelog"
)

func main()  {

        logger, err := seelog.LoggerFromConfigAsFile("/home/winter/go/seelog.xml")
        if err != nil {
                panic(err)
        }
        seelog.ReplaceLogger(logger)
        defer seelog.Flush()
        seelog.Debug("hello world! 1")
        seelog.Debug("hello world! 2")
        seelog.Debug("hello world! 3")
        seelog.Debug("hello world! 4")
        seelog.Debug("hello world! 5")
        seelog.Debug("hello world! 6")
        seelog.Debug("hello world! 7")
        seelog.Debug("hello world! 8")
        seelog.Debug("hello world! 9")
        seelog.Error("hello world! 1")
        seelog.Error("hello world! 2")
        seelog.Error("hello world! 3")
        seelog.Error("hello world! 4")
        seelog.Error("hello world! 5")
        seelog.Error("hello world! 6")
        seelog.Error("hello world! 7")
        seelog.Error("hello world! 8")
        seelog.Error("hello world! 9")
//      seelog.Flush()
}
