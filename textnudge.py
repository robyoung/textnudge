import os
from flask import Flask, request
from twilio.rest import TwilioRestClient
from twilio import twiml

app = Flask(__name__)
app.config['PERSON_ONE'] = os.getenv('PERSON_ONE')
app.config['PERSON_TWO'] = os.getenv('PERSON_TWO')
client = TwilioRestClient()


@app.route('/')
def hello():
    return "Hello world!"


@app.route('/receive', methods=['POST'])
def receive():
    print("{} {} {}".format(request.method, request.data, request.form))
    message = client.messages.create(
        to=get_forward_number(request.form),
        from_=request.form['To'],
        body=request.form['Body'])
    print("{}".format(message))

    r = twiml.Response()
    r.message("You are the bestest!")
    return str(r)


def get_forward_number(form):
    if form['From'] == app.config['PERSON_ONE']:
        return app.config['PERSON_TWO']
    elif form['From'] == app.config['PERSON_TWO']:
        return app.config['PERSON_ONE']
    else:
        raise ValueError("Invalid from address {}".format(form['From']))

if __name__ == '__main__':
    app.debug = True
    app.run()
