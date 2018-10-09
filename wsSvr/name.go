package main

import "math/rand"

func createName() string {
	nickname := []string{"呼保义", "玉麒麟", "智多星", "入云龙", "大刀", "豹子头", "霹雳火", "双鞭", "小李广", "小旋风", "扑天雕", "美 髯公", "花和尚", "行者", "双枪将", "没羽箭", "青面獣", "金枪手", "急先锋", "神行太保", "赤髪鬼", "黒旋风", " 九纹龙", "没遮拦", "挿翅虎", "混江龙", "立地太歳", "船火児", "短命二郎", "浪里白跳", "活阎罗", "病关索", "拚命三郎", "两头蛇", "双尾蝎", "浪子", "星名", "神机军师", "镇三山", "病尉遅", "丑郡马", "井木犴", "百胜将", " 天目将", "圣水将", "神火将", "圣手书生", "鉄面孔目", "摩云金翅", "火眼狻猊", "锦毛虎", "锦豹子", "轰天雷", "神算子", "小温侯", "赛仁贵", "神医", "紫髯伯", "矮脚虎", "一丈青", "丧门神", "混世魔王", "毛头星", "独火星", "八臂哪吒", "飞天大圣", "玉臂匠", "鉄笛仙", "出洞蛟", "翻江蜃", "玉幡竿", "通臂猿", "跳涧虎", "白花蛇", "白 面郎君", "九尾亀", "鉄扇子", "鉄叫子", "花项虎", "中箭虎", "小遮拦", "操刀鬼", "云里金刚", "摸着天", "病大虫", "打虎将", "小霸王", "金銭豹子", "鬼睑児", "出林龙", "独角龙", "旱地忽律", "笑面虎", "金眼彪", "鉄臂膊", " 一枝花", "催命判官", "青眼虎", "没面目", "石将军", "小尉遅", "母大虫", "菜园子", "母夜叉", "活闪婆", "険道神", "白日鼠", "鼓上蚤", "金毛犬"}
	n := rand.Intn(len(nickname))
	return nickname[n]
}
