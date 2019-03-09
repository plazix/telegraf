# Temp (from /sys/class/thermal/thermal_zone*) input plugin

The temp input plugin gather metrics on system temperature.  This plugin is
meant to be multi platform and uses platform specific collection methods.

Currently supports only Linux

### Configuration:

```
[[inputs.thermal_zone]]
```
