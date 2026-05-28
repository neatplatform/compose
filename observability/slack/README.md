# slack

Small HTTP service for experimenting with Slack app callbacks and interaction flows.

It exposes Slack-compatible endpoints for:

  - Slash commands at `/commands`
  - Events API callbacks at `/events`
  - Interactive payloads at `/interactions`
  - Select menu payloads at `/options`

It also exposes:

  - Prometheus metrics at `/metrics`
  - pprof endpoints at `/debug/pprof`

## What It Does

At startup, the service posts a *Block Kit* fixture message to Slack, then starts an HTTP server.
From there, it handles common callback types:

  - `/commands`: acknowledges slash commands and posts follow-up responses.
  - `/events`: supports Slack URL verification and event callbacks.
    - `message` events: adds a `:thumbsup:` reaction.
    - `app_mention` events: replies in thread with `Noted!`.
  - `/interactions`: handles interactive components.
    - `block_actions`: can open a modal for the fixture button action.
    - `shortcut`: parsed and logged.
    - `message_action`: parsed and logged.
    - `view_submission`: parsed and logged.
    - `view_closed`: parsed and logged.
  - `/options`: handles options-load callbacks for select menus.
    - `block_suggestion`: responds with options and/or option groups.

All Slack callbacks are verified using Slack request **signatures** and **timestamps**.

## Create a New Slack App

Create a new Slack app:

  1. Go to https://api.slack.com/apps.
  2. Click "Create New App", then choose "From a manifest".
  3. Select a workspace and click on "Next".
  4. Copy the contents of `app.manifest.yaml`, paste them into the YAML editor, and click "Next".
  5. Review the OAuth scopes, then click "Create".

Once your app is created:

  - Install it into your workspace.
  - Add it to the channels you want to use.

## Quick Start

Required environment variables:

  - `NGROK_AUTHTOKEN`: ngrok auth token
  - `SLACK_APP_SIGNING_SECRET`: Slack app signing secret (*Slack App Settings -> Basic Information*)
  - `SLACK_APP_AUTH_TOKEN`: OAuth token (*Slack App Settings -> OAuth & Permissions*)

Build and run locally:

```bash
make
./slack -h
```

## Resources

  - [Slack API in Go](https://github.com/slack-go/slack)
  - [Block Kit Builder](https://app.slack.com/block-kit-builder)
  - **Docs**
    - [App design](https://docs.slack.dev/concepts/app-design)
    - [Creating an app from app settings](https://docs.slack.dev/app-management/quickstart-app-settings)
    - **API**
      - [Events API](https://docs.slack.dev/apis/events-api)
        - [Using HTTP Request URLs](https://docs.slack.dev/apis/events-api/using-http-request-urls)
        - [Using Socket Mode](https://docs.slack.dev/apis/events-api/using-socket-mode)
    - **Authentication**
      - [Using token rotation](https://docs.slack.dev/authentication/using-token-rotation)
      - [Using PKCE](https://docs.slack.dev/authentication/using-pkce)
      - [Verifying requests from Slack](https://docs.slack.dev/authentication/verifying-requests-from-slack)
    - [**Interactivity**](https://docs.slack.dev/interactivity)
      - [Handling user interaction in your Slack apps](https://docs.slack.dev/interactivity/handling-user-interaction)
      - [Implementing shortcuts](https://docs.slack.dev/interactivity/implementing-shortcuts)
      - [Implementing slash commands](https://docs.slack.dev/interactivity/implementing-slash-commands)
    - [**Messaging**](https://docs.slack.dev/messaging)
      - [Creating interactive messages](https://docs.slack.dev/messaging/creating-interactive-messages)
      - [Formatting message text](https://docs.slack.dev/messaging/formatting-message-text)
    - [**Block Kit**](https://docs.slack.dev/block-kit)
      - [Formatting with rich text](https://docs.slack.dev/block-kit/formatting-with-rich-text)
    - **Reference**
      - [Block Kit](https://docs.slack.dev/reference/block-kit)
      - [Events](https://docs.slack.dev/reference/events)
      - [Interaction payloads](https://docs.slack.dev/reference/interaction-payloads)
      - [Methods](https://docs.slack.dev/reference/methods)
