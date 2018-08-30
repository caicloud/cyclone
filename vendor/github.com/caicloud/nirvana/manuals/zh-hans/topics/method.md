# method 包


在 definition 包中，可以看到 Definition.Function 必须是一个函数，这就要求业务逻辑必须放在函数中，而不能在实例方法中。为了解决这个问题，method 包提供了一个实例方法容器，帮助用户把实例方法转换为函数。

容器如下：
```go
// Container contains instances and mappings.
type Container struct {
	...
}

// Put puts an instance in this container. The instance must have one or more methods.
func (c *Container) Put(instance interface{})

// PutInterface puts an instance in this container. The instance must have one or more methods.
// The iface should be like (*ArbitraryInterface)(nil).
func (c *Container) PutInterface(iface interface{}, instance interface{})

// Get returns a function for specified method. If you want to specify a method from an
// interface, you need to use (*ArbitraryInterface)(nil) as instance.
func (c *Container) Get(instance interface{}, method string) interface{}
```

这个实例方法容器分离了方法的 Get 和 Put 过程。也就是可以在声明 API 时，使用 Get 获取某个实例的方法，之后再在服务启动逻辑里 Put 真正的实例，即 Get 可以在 Put 之前使用。

这个主要是利用了 golang 的动态函数构建的功能，先生成函数，但是在函数被真正调用的时候才去获取真正的实例里的方法。

method 包还提供了一个全局的 Container：
```go
var defaultContainer = NewContainer()

// Put puts an instance in this container. The instance must have one or more methods.
func Put(instance interface{}) {
	defaultContainer.Put(instance)
}

// PutInterface puts an instance in this container. The instance must have one or more methods.
// The iface should be like (*ArbitraryInterface)(nil).
func PutInterface(iface interface{}, instance interface{}) {
	defaultContainer.PutInterface(iface, instance)
}

// Get returns a function for specified method. If you want to specify a method from an
// interface, you need to use (*ArbitraryInterface)(nil) as instance.
func Get(instance interface{}, method string) interface{} {
	return defaultContainer.Get(instance, method)
}
```
一般情况下，用户会使用这个全局的容器。
