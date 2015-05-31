import os
from datetime import datetime

from celery import Celery
from dateutil.parser import parse as parse_date

from .web import redis_client, twilio_client

app = Celery('tasks')
app.conf.update(
    BROKER_URL=os.getenv('REDISCLOUD_URL'),
    CELERY_TASK_SERIALIZER='json',
    CELERY_ACCEPT_CONTENT=['json'],  # Ignore other content
    CELERY_RESULT_SERIALIZER='json',
    CELERYD_CONCURRENCY=1,
)


@app.task
def add(x, y):
    return x + y


@app.task
def nudge(to_number, twilio_number):
    key = "textnudge.unreplied.{}".format(to_number)
    length = redis_client.llen(key)
    if length > 0:
        oldest = parse_date(redis_client.lindex(key, length - 1))
        message = twilio_client.messages.create(
            to=to_number,
            from_=twilio_number,
            body="You have unreplied to messages. Oldest is {}".format(
                datetime.now() - oldest))
        nudge.apply_async(args=[to_number, twilio_number], countdown=300)
        # TODO: handle errors
        print("{}".format(message))
