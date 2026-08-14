The app should have ai integration to be able to use an ai agent for chat requests.

Extend the config package to load an auth file from ~/.config/djinni/config.yaml

The config should contain information about model providers that can be used for ai.
Example:

modelProviders:
- name: "litellm"
  apiBase: "https://mylitellm"
  apiKey: "someapikey"
  models:
  - id: "mymodel"

Use langchainGo for implementing an agent. The agent should have the possibility to read files 
of specific folders and to execute readonly git commands in a folder of choice. 