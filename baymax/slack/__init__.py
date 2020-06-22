from flask import Blueprint

import os
import logging
from slack import RTMClient
from slack.errors import SlackApiError

baymax_slack_rtm = Blueprint(
    name="baymax_slack_rtm",
    import_name=__name__,
)


@RTMClient.run_on(event='message')
def handle_message(**payload):
    print(payload)

@RTMClient.run_on(event='app_mention')
def handle_mention(**payload):
    print(payload)

rtm_client = RTMClient(token=os.environ["SLACK_API_TOKEN"])
rtm_client.start()