# 过期删除策略

* 定期删除
* 惰性删除

# 内存淘汰策略

* noeviction:返回错误，不会删除任何键值.
* allkeys-lru:使用LRU算法删除最近最少使用的键值
* volatile-lru:使用LRU算法从设置了过期时间的键集合中删除最近最少使用的键值
* allkeys-random:从所有key随机删除
* volatile-random:从设置了过期时间的键的集合中随机删除
* volatile-ttl:从设置了过期时间的键中删除剩余时间最短的键
* volatile-lfu:从配置了过期时间的键中删除使用频率最少的键
* allkeys-lfu:从所有键中删除使用频率最少的键