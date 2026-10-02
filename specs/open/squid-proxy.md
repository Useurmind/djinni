# Internal Network and Proxy

We want the agent container to be isolated from the internet to ensure no illegal internet access is performed.

Therefore it should be hosted in an internal network.
The name of that network should derived from the agent that is hosted in it.

The config however should allow internet access if the user desires this.

    agents:
      default:
        network:
          internal: true  # default false, if true agent network is internal
          proxy:
            enabled: true # default false, if true the proxy is created and the agent can access it
            allowList:    # default no entries, these are configured into the proxy as an acl
            - github.com
            - google.com

This is done by creating a second proxy container, that is attached to the internal network and the default network (for internet access).
In this way the agent container can only access the internet by traversing the proxy.

The following components are needed

    Container 1:
        OpenCode

    Container 2:
        Squid

    Network:
        ai-net (internal)

    Squid:
        ai-net + podman

    OpenCode:
        ai-net only
