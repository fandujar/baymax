from flask import Flask
from baymax.plugins import baymax_plugins
from baymax.slack import baymax_slack_rtm

app = Flask(__name__)
app.register_blueprint(baymax_slack_rtm)
#app.register_blueprint(baymax_plugins)


app.run(
    host="0.0.0.0",
    port=8080
)

