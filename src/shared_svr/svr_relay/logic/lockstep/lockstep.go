/***********************************************************************
* @ 帧同步
* @ brief
	· 忘掉时间，游戏一帧一帧进行的，预表现层掩盖卡顿
		· 逻辑层
			· 严格的回合制
			· 碰撞等的判定，取逻辑层数据
		· 预表现层
			· 平滑过渡逻辑数据
			· 反馈慢的游戏（野蛮人大作战），可不要预表现层，等收到服务器指令集，本地再执行

    · 服务器保证：大家每帧收到的输入一致
		· 客户端收到的指令集，必须被执行（收到指令时，上下文未准备好）
		· 比如场景未初始化完，就收到指令了，须缓存

	· 客户端保证：输入一致，输出也一致
		· 浮点数改定点数
		· 确定性排序
		· 统一随机数
		· 动画timing事件，帧改造
		· 定时器，帧改造
		· 物理引擎
		· AI

* @ author zhoumf
* @ date 2020-5-15
***********************************************************************/
package lockstep

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"shared_svr/svr_relay/player"
)

// Input：玩家的操作输入数据(按键等状态的全量)，客户端每50ms上报一次
// Frame：一帧中各玩家的Input集合
func Rpc_relay_report_input(req, ack *common.NetPack, this *player.Player) {
	oplog := req.ReadString() //每个bit代表一种操作
	state := req.ReadUInt32() //客户端状态hash
	G_test.AddInput(this.Pid, oplog, state)
}
func Rpc_relay_lockstep_begin(req, ack *common.NetPack, this *player.Player) {
	G_test.Clear()
}
func Rpc_relay_lockstep_end(req, ack *common.NetPack, this *player.Player) {
	G_test.Clear()
}

// ------------------------------------------------------------
type TInput struct {
	Pid   uint32
	Oplog string //每个bit代表一种操作
	State uint32 //客户端状态hash
}
type TFrame []TInput

type Mgr struct {
	logs  []TFrame //玩家操作记录
	frame TFrame
}

var G_test Mgr

func (m *Mgr) Clear() {
	m.logs = m.logs[:0]
	m.frame = m.frame[:0]
}
func (m *Mgr) AddInput(pid uint32, oplog string, state uint32) {
	m.frame = append(m.frame, TInput{pid, oplog, state})
}
func (m *Mgr) Broadcast() {
	if len(m.frame) == 0 {
		return
	}
	m.logs = append(m.logs, m.frame)
	for i := 0; i < len(m.frame); i++ {
		if v := player.FindPlayer(m.frame[i].Pid); v != nil {
			v.Conn.CallRpc(enum.Rpc_client_handle_spawn_attrs, func(buf *common.NetPack) {
				buf.WriteInt(len(m.logs)) //帧编号
				m.frame.writeTo(buf)
			}, nil)
		} else {
			gamelog.Error("Broadcast: pid(%d)", m.frame[i].Pid)
		}
	}
	m.frame = m.frame[:0]
}
func (f TFrame) writeTo(buf *common.NetPack) {
	for i := 0; i < len(f); i++ {
		buf.WriteUInt32(f[i].Pid)
		buf.WriteString(f[i].Oplog)
		// 比对客户端状态hash
		for j := i + 1; j < len(f); j++ {
			if f[i].State != f[j].State {
				gamelog.Error("State diff: %v", f)
			}
		}
	}
}

// ------------------------------------------------------------
// 状态hash
