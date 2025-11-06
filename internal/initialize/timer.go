package initialize

import (
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils/timer"

	"github.com/robfig/cron/v3"
)

func Timer(serviceCtx *svc.ServiceContext) {
	serviceCtx.Timer = timer.NewTimerTask()
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())

		// 其他定时任务定在这里 参考上方使用方法

		//_, err := global.GVA_Timer.AddTaskByFunc("定时任务标识", "corn表达式", func() {
		//	具体执行内容...
		//  ......
		//}, option...)
		//if err != nil {
		//	fmt.Println("add timer error:", err)
		//}
	}()
}
