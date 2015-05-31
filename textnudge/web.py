import os
from datetime import datetime

from flask import Flask, request
from twilio.rest import TwilioRestClient
from twilio import twiml

from .utils import connect_redis, get_forward_number
from .tasks import nudge

app = Flask(__name__)
app.config['PERSON_ONE'] = os.getenv('PERSON_ONE')
app.config['PERSON_TWO'] = os.getenv('PERSON_TWO')
twilio_client = TwilioRestClient()
redis_client = connect_redis(os.getenv('REDISCLOUD_URL'))


@app.route('/')
def hello():
    return "Hello world!"


@app.route('/receive', methods=['POST'])
def receive():
    to_number = get_forward_number(request.form['From'], app.config)
    from_number = request.form['From']
    twilio_number = request.form['To']

    message = twilio_client.messages.create(
        to=to_number,
        from_=twilio_number,
        body=request.form['Body'])
    print("{}".format(message))

    redis_client.lpush("textnudge.unreplied.{}".format(to_number),
                       datetime.now().isoformat())
    redis_client.delete("textnudge.unreplied.{}".format(from_number))

    nudge.apply_async(args=[to_number, twilio_number], countdown=300)

    r = twiml.Response()
    return str(r)


if __name__ == '__main__':
    app.debug = True
    app.run()
