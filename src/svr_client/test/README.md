https://github.com/idada/go-labs

一些杂七杂八脑洞大开的实验：

| 实验 | 描述 |
| ------ | ------ |
| labs01 | 测试类型判断和类型转换的效率 |
执行结果：
    dada-imac:misc dada$ go test -test.bench=".*" labs01
    testing: warning: no tests to run
    PASS
    Benchmark_TypeSwitch         50000000            33.0 ns/op
    Benchmark_NormalSwitch       2000000000          1.99 ns/op
    Benchmark_InterfaceSwitch    100000000           18.4 ns/op
    ok      labs    7.741s
结论：类型判断和类型转换这两个操作都比直接操作多几倍的消耗。


| labs03 | 测试对象创建的效率 |
实验结果：
    dada-imac:misc dada$ go test -test.bench="." labs03
    testing: warning: no tests to run
    PASS
    Benchmark_NewStruct1    100000000            13.0 ns/op     //return new(BigStruct)
    Benchmark_NewStruct2    100000000            24.5 ns/op     //return &BigStruct{}
    Benchmark_NewStruct4    100000000            25.1 ns/op     //*r = BigStruct{}      //外部new指针传入，这里实际有临时对象拷贝、销毁
    Benchmark_NewStruct5    200000000            8.65 ns/op     //r.field = 0           //外部new指针传入，只编辑了对象，实际仅仅创建了一个
    ok      labs03  10.872s
结论：5和4的结果比较意外。


| labs04 | 测试range循环和for循环，以及结构体循环和指针循环的性能区别 |
实验结果：
    dada-imac:misc dada$ go test -test.bench="." labs04
    testing: warning: no tests to run
    PASS
    Benchmark_Loop1     2000000         923 ns/op
    Benchmark_Loop2     2000000         819 ns/op
    Benchmark_Loop3     2000000         825 ns/op
    Benchmark_Loop4     100000        26230 ns/op
    ok      labs04  10.640s
结论：对结构体列表的range循环最消耗性能，因为数据要重复复制。


| labs06 | 测试小数据量循环和map取数据以及硬编码取数据的性能消耗 |
试验结果：
    dada-imac:labs dada$ go test -test.bench="." labs06
    testing: warning: no tests to run
    PASS
    Benchmark_Loop1    500000000             5.73 ns/op
    Benchmark_Loop2    500000000             5.72 ns/op
    Benchmark_Loop3    50000000              68.0 ns/op
    Benchmark_Loop4    500000000             4.92 ns/op
    Benchmark_Loop5    500000000             4.40 ns/op
    ok      labs06  15.970s
结论：硬编码 < 指针slice的range循环 < for循环，但是量级是一样的，看情况用。但是map差了一个量级，小数据量尽量少用。


| labs09 | 测试匿名函数和普通函数的调用消耗 |
禁用优化前：
    dada-imac:labs dada$ go test -test.bench="." labs09
    testing: warning: no tests to run
    PASS
    Benchmark_NormalFuncCall  2000000000         0.52 ns/op
    Benchmark_VarFuncCall1    2000000000         1.84 ns/op
    Benchmark_ClosureCall2    1000000000         2.16 ns/op
    ok      labs09  7.347s
禁用优化后：
    dada-imac:labs dada$ go test -test.bench="." -gcflags '-N' labs09
    testing: warning: no tests to run
    PASS
    Benchmark_NormalFuncCall  2000000000         1.99 ns/op
    Benchmark_VarFuncCall1    1000000000         2.02 ns/op
    Benchmark_ClosureCall2    50000000           58.8 ns/op
    ok      labs09  9.412s
结论：
    测试中用的是一个很简单的函数，所以很容易被新版的go做内联优化，禁用优化前后结果差别明显。
    如果不考虑优化，匿名函数和普通函数调用是一个量级的消耗，差别甚微。但是普通函数比较有可能被优化。
    在没有优化的情况下，闭包函数消耗又要再高一个量级。


| labs12 | 测试jsmalloc和malloc在go程序中是否有性能差别 |
实验结果：
    dada-imac:labs dada$ go test -test.bench=".*" labs12
    PASS
    Benchmark_AllocAndFree  10000000           190 ns/op
    Benchmark_JeAllocAndFree    20000000           135 ns/op
    ok      labs12  4.971s
结论：还是有差的


| labs13 | 测试不同数据结构的对象数量差别 |
存结构体和存指针的map类型占用的对象数量：
    go run many_object1.go -gcflags '-N' | grep -o "HeapObjects.*"
    go run many_object2.go -gcflags '-N' | grep -o "HeapObjects.*"

存结构体和存指针的slice类型占用的对象数量：
    go run many_object3.go -gcflags '-N' | grep -o "HeapObjects.*"
    go run many_object4.go -gcflags '-N' | grep -o "HeapObjects.*"

未初始化的Slice和make后的Slice类型字段占用的对象数量：
    go run many_object5.go -gcflags '-N' | grep -o "HeapObjects.*"  //data := BigStruct{}
    go run many_object6.go -gcflags '-N' | grep -o "HeapObjects.*"  //data := BigStruct{ make([]int, 0, 1) }
结论：
    1. 存指针的map和存结构体的map，每条数据都占用一个对象数量，两种数据结构无差异
    2. 存指针的slice和存结构体的slice有明显差异，存结构体的slice不重复占用对象数量
    3. 未初始化的对象类型字段是不占用对象数量的


| labs24 | 测试binary.Write和硬编码的效率差异 |
测试结果：
    dada-imac:labs24 dada$ go test -v -bench='.*'
    testing: warning: no tests to run
    PASS
    Benchmark_UseBinaryWrite1   50000000        68.5 ns/op  //i := 0; UseBinaryWrite1(w, int32(i))
    Benchmark_UseBinaryWrite2   10000000        279 ns/op   //i := 0; UseBinaryWrite1(w, i)
    Benchmark_UseHardcode       100000000       21.5 ns/op  //var i int32 = 0; io.Write([]byte{ byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24) })
    ok      github.com/idada/go-labs/labs24 8.771s


| labs25 | 测试LockOSThread()对chan消息处理的影响 |
测试结果：
    $ go test -bench="."
    testing: warning: no tests to run
    PASS
    Benchmark_Normal        10000000        220 ns/op
    Benchmark_LockThread     1000000        2084 ns/op
    ok      github.com/idada/go-labs/labs25 4.540s
结论：
    本来是希望通过LockOSThread()可以独占线程，避免chan通讯发生的调度开销。
    结果跟预期相反，由于调度算法没有改变。
    LockOSThread()的goroutine反而需要等待关联的线程空闲了才能被执行到，所以反而是变慢了。


| labs26 | 比较直接调用函数和反射调用函数的效率差别 |
测试结果：
    $ go test -bench="."
    testing: warning: no tests to run
    PASS
    Benchmark_NormalFuncCall    2000000000         0.53 ns/op
    Benchmark_ReflectFuncCall    2000000           667 ns/op    //多了三次额外的参数传递反射
    ok      github.com/idada/go-labs/labs26 3.116s


| labs28 | 测试`[]byte`转`string`的效率 |
测试结果：
    $ go test -bench="."
    PASS                                                // var x = []byte("Hello World!")
    Benchmark_Normal        20000000        63.4 ns/op  // _ = string(x)
    Benchmark_ByteString    2000000000      0.55 ns/op  // _ = *(*string)(unsafe.Pointer(&x))
    ok      github.com/idada/go-labs/labs28 2.486s


| labs17 | 尝试优化一段代码 |
| labs30 | 内存数据库事务Demo |
| labs31 | 计算一个值在2的第N次幂区间（蛋疼的性能测试）|