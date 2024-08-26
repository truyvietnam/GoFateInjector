# GoFateInjector

A sh*tty replica of Fate Injector written in Go because why not!

**Note: This injector is not associated with bigrat.monster website. I put bigrat.monster in "Made by bigrat.monster" as a joke**

# Why

1. Cuz I am a crackhead
2. I want to make the first Go injector that have gui (public) so I choose to replica Fate Injector with Win32 gui from Windigo!
3. Fate Injector doesn't support high dpi (I am using 125% scale) so it look blurry
   
![blurry](https://raw.githubusercontent.com/truyvietnam/GoFateInjector/main/image/fate%20injector%20125scale%20blur.png)

![notscaled](https://github.com/truyvietnam/GoFateInjector/blob/main/image/fate%20injector%20125scale%20problem.png?raw=true)

# Images

The images was shot at 125% scale

![default](https://github.com/truyvietnam/GoFateInjector/blob/main/image/GoFateInjector_default.png?raw=true)

![selectDll](https://github.com/truyvietnam/GoFateInjector/blob/main/image/GoFateInjector_selectdll.png?raw=true)

![delayBox](https://github.com/truyvietnam/GoFateInjector/blob/main/image/NumberOnly.png?raw=true)

# 100% scale temporary solution

**If you are using higher scale you don't have to follow this**

There is a problem with status bar on 100% scale that looks higher than normal and it will look annoy

![problem](https://github.com/truyvietnam/GoFateInjector/blob/main/image/GoFateInjector_100scale.png?raw=true)

This may be the problem happened with windigo. To temporarily fix this you just need to set the injector's high DPI setting by right-click the exe -> Properties -> Compatibility tab -> Change high DPI settings with this following setting and how the injector will look like

![setting](https://github.com/truyvietnam/GoFateInjector/blob/main/image/scale100fix.png?raw=true)

# Credits
[Fate Injector](https://github.com/fligger/FateInjector) - The original Fate Injector

[Windigo](https://github.com/rodrigocfd/windigo) - Win32 gui library in Go

[injgo](https://github.com/jiusanzhou/injgo) - LoadLibrary inject method
