# Koru metrics solution
Monitoring performance is key, therefore this package is created to assist working with Koru3D.

### How it works
Monitoring is always disabled by default, to enable it, use the `-metrics` flag and specify the method after.
Available methods are:
- [ ] `simplejson` - SimpleJson endpoint for use with Grafana


### Notes
When SimpleJson is selected metrics are held in memory, therefore an arbitrary max-holding time of 1 hour is present.
Metrics after that time will be discarded.