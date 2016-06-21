consul-notify
-------------

The purpose of this project is to monitor consul and send notifications to various plugins. 
This project was based on [consul-alerts](https://github.com/AcalephStorage/consul-alerts)
consul-alerts is very feature rich and complicated. consul-notify seeks to address this complexity by only doing one thing, forwarding alerts to notifiers.

usage
-----

consul-notify should initially be run in the start mode. In turn, the start mode will run consul watch and feed the output to the consul-notify watch mode.

    Consul Notify.
    
    Usage:
      consul-notify start [--config=<config>]
      consul-notify watch [--config=<config>]
      consul-notify --help
      consul-notify --version
    
    Options:
      --config=<config>            The consul-notify config [default: /etc/consul-notify/consul-notify.conf].
      --help                       Show this screen.
      --version                    Show version.

base configuration
------------------

The consul configuration should point to a consul service. These are the defaults:

    [consul]
      addr = "localhost:8500"
      dc = "dc1"
  
alerta
------

    [alerta]
      url = "http://localhost:8000"
      token = ""