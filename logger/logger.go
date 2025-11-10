package logger

import (
    "github.com/sirupsen/logrus"
    "os"
)


func Init() {
    logrus.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })
    logrus.SetOutput(os.Stdout)
    logrus.SetLevel(logrus.DebugLevel)
}
