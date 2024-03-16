package VixDiskLibApi

// 定义日志回调方法

/*
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>

void LogFunc (const char* fmt, va_list args)
{
    char logBuff[4096];
    vsnprintf(logBuff, sizeof(logBuff), fmt, args);
    GoLogFunc(logBuff);
}

void WarnFunc (const char* fmt, va_list args)
{
    char logBuff[4096];
    vsnprintf(logBuff, sizeof(logBuff), fmt, args);
    GoWarnFunc(logBuff);
}

void PanicFunc (const char* fmt, va_list args)
{
    char logBuff[4096];
    vsnprintf(logBuff, sizeof(logBuff), fmt, args);
    GoPanicFunc(logBuff);
}
*/
import "C"
