/***********************************************************************
* @ jump consistent hash
* @ brief
	1、输入是一个64位的key，和桶的数量（一般对应服务器的数量），输出桶编号
	2、业务层记录 <key, idx>
	3、桶数变化时，遍历记录重新计算idx，只需迁移idx变动过的

* @ author zhoumf
* @ date 2019-3-15
***********************************************************************/
package hash

func JumpHash(key uint64, capBuckets int) int32 {
	b, j := int64(-1), int64(0)
	for j < int64(capBuckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}
	return int32(b)
}
