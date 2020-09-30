package main

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"github.com/zsdevX/DarkEye/common"
	"github.com/zsdevX/DarkEye/fofa"
	"strconv"
	"time"
)

var (
	//fofa功能界面
	windowFofa = &widgets.QMainWindow{}
)

func registerFofa() (window *widgets.QMainWindow) {
	window = widgets.NewQMainWindow(nil, 0)
	//Input
	ip := widgets.NewQLineEdit(nil)
	ip.SetPlaceholderText("IP（Nmap格式）")
	ip.SetToolTip("1.1.1.1-254,2.2.2.2")

	Interval := widgets.NewQLineEdit(nil)
	Interval.SetToolTip("检索间隔（建议10秒）")
	Interval.SetText("10")
	Interval.SetAlignment(core.Qt__AlignHCenter)

	session := widgets.NewQLineEdit(nil)
	session.SetPlaceholderText("_fofapro_ars_session=xxx")
	session.SetToolTip("不填写仅能获取一页fofa记录")
	session.SetAlignment(core.Qt__AlignHCenter)
	if mConfig.Fofa.FofaSession != "" {
		session.SetText(mConfig.Fofa.FofaSession)
	}

	//Log
	logC, runCtl, inputLog := getWindowCtl()

	btnOpen := widgets.NewQPushButton2("启动", nil)
	btnStop := widgets.NewQPushButton2("停止", nil)
	btnStop.SetDisabled(true)

	widgetC := widgets.NewQWidget(nil, 0)
	widgetC.SetLayout(widgets.NewQHBoxLayout())
	widgetC.Layout().AddWidget(ip)

	widgetD := widgets.NewQWidget(nil, 0)
	widgetD.SetLayout(widgets.NewQHBoxLayout())
	widgetD.Layout().AddWidget(Interval)
	widgetD.Layout().AddWidget(session)

	widgetA := widgets.NewQWidget(nil, 0)
	widgetA.SetLayout(widgets.NewQVBoxLayout())
	widgetA.Layout().AddWidget(inputLog)

	widgetB := widgets.NewQWidget(nil, 0)
	widgetB.SetLayout(widgets.NewQHBoxLayout())
	widgetB.Layout().AddWidget(btnOpen)
	widgetB.Layout().AddWidget(btnStop)

	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	widget.Layout().AddWidget(widgetC)
	widget.Layout().AddWidget(widgetD)
	widget.Layout().AddWidget(widgetA)
	widget.Layout().AddWidget(widgetB)

	window.SetMinimumSize2(650, 480)
	window.SetWindowTitle(programDesc)
	window.SetCentralWidget(widget)
	window.AutoFillBackground()
	window.SetWindowFlags(core.Qt__Dialog)

	//Action
	btnOpen.ConnectClicked(func(bool) {
		//保存配置
		mConfig.Fofa = fofa.NewConfig()
		mConfig.Fofa.Ip = ip.Text()
		mConfig.Fofa.Interval, _ = strconv.Atoi(Interval.Text())
		mConfig.Fofa.FofaSession = session.Text()
		mConfig.Fofa.ErrChannel = logC
		if err := saveCfg(); err != nil {
			logC <- common.LogBuild("UI", "保存配置失败:"+err.Error(), common.FAULT)
			return
		}
		//启动流程
		common.StartIt(&mConfig.Fofa.Stop)
		go func() {
			mConfig.Fofa.Run()
			btnStop.SetDisabled(true)
			btnOpen.SetDisabled(false)
			runCtl <- false
		}()
		btnStop.SetDisabled(false)
		btnOpen.SetDisabled(true)
		window.SetWindowState(core.Qt__WindowNoState)
	})

	btnStop.ConnectClicked(func(bool) {
		common.StopIt(&mConfig.Fofa.Stop)
		btnStop.SetDisabled(true)
		//异步处理等待结束避免界面卡顿
		go func() {
			sec := 0
			stop := false
			tick := time.NewTicker(time.Second)
			for {
				select {
				case <-runCtl:
					stop = true
				case <-tick.C:
					sec ++
					btnStop.SetText(fmt.Sprintf("等待%d秒", 60-sec))
				}
				if stop {
					break
				}
			}
			btnOpen.SetDisabled(false)
			btnStop.SetText("停止")
		}()
	})
	return
}
