-   groupcache 代码粗读

-   简介

groupcache is a distributed caching and cache-filling library, intended as a replacement for a pool of memcached nodes in many cases.

-   特点

1. 只是一个库，可以集成到现有服务上，也可以自己写一个单独的服务。对访问的用户来说，它是服务端，对其他 peer 来说它又是客户端

2. 缓冲填充机制（singleflight），解决缓冲击穿的问题，多个同时的请求在本地缓冲未命中的情况下只有一个会去请求 peer 或数据库

3. 不能更新数据， 不能删除数据！非常限制 groupcache 的使用场景，导致它基本只能用在数据不变的场景（immutable），比如时序性数据。（本人思考了一下添加更新或者删除功能的可能性，发现 groupcache 的设计和架构导致它很难有更新或者删除的功能。（分布式架构下）更新很多时候不能在读的时候知晓，所以读的时候强制更新没办法确定使用时机，而在写的时候更新或删除涉及到 hotcache，localcache,peercache 的一致性问题，比较麻烦，需要使用者自己维护所有的 peer 数组，通知所有 peer）

4. 支持自动本地保存热点数据，hotcache（未完成，只是随机保存一些数据）

5. 只支持 go

6. 底层使用 lru，意味着使用中基本会保持占用设定的最大内存，数据淘汰只能靠 lru 的机制

7. 分布式，使用一致性 hash

-   分析

本文自底向上进行介绍吧

1.  底层的缓存储存使用 lru，达到 max size 后自动淘汰最早使用的数据，单独 lru 没有线程安全，增删查改都是 o(1)，具体介绍看我对 lru 的解析

2.  分布式储存使用一致性 hash，保证 peer 上下线时最少的缓存失效，具体介绍看另一个文章

3.  使用 singleflight 解决缓存击穿，具体介绍看另一个文章

4.  储存格式在 byteview，使用 byte 数组或者 string 进行储存

5.  主要逻辑都在 groupcache 中

        cache struct 实现了lru的线程安全和lru的使用统计

        Group struct 维护了本地cache, hotcache, PeerPicker, getter(缓存未命中的获取)函数
        主要方法：
        NewGroup(name string, cacheBytes int64, getter Getter) *Group
        Get(ctx Context, key string, dest Sink) 先尝试本地获取(maincache,hotcache),之后singleflight中依次尝试本地，peers，getter获取

6.  http 实现了本地服务，以及本地和 peers 的通信，同时维护 peers。http 和 groupcache 通过 peers.go 中的全局变量实现了解耦(http 可以直接调用 groupcache，groupcache 使用 peers 中的全局变量，http 注册 peers 中的全局变量)，这样可以自己重新实现 http 中的通信协议，自己维护 peers。
    Set(peers ...string)
    Get(context Context, in *pb.GetRequest, out *pb.GetResponse)
