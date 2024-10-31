package gozen

import (
	"errors"
	"fmt"
	"sync"
)

func GoFuncOne(f func() error) {
	defer func() {
		if r := recover(); r != nil {
			LogErrorw(LogNameLogic, "GoFuncOne panic",
				LogKNameCommonErr, r,
				LogKNameCommonKey, RunFuncNameSkip(4),
			)
		}
	}()
	err := f()
	if err != nil {
		LogErrorw(LogNameLogic, "GoFuncOne error",
			LogKNameCommonErr, err,
		)
	}
}

func GoFunc(f ...func() error) (err error) {
	var wg sync.WaitGroup
	for _, v := range f {
		wg.Add(1)
		go func(handler func() error) {
			defer func() {
				if r := recover(); r != nil {
					LogErrorw(LogNameLogic, "GoFunc panic",
						LogKNameCommonErr, r,
						LogKNameCommonKey, RunFuncNameSkip(4),
					)
					err = errors.New("panic:" + fmt.Sprint(r)) //某一个func panic后，报错
				}
				wg.Done()
			}()
			e := handler()
			if err == nil && e != nil {
				err = e
			}
		}(v)
	}
	wg.Wait()
	return
}
