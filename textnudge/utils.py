from urllib.parse import urlparse
import redis


def get_forward_number(from_, config):
    if from_ == config['PERSON_ONE']:
        return config['PERSON_TWO']
    elif from_ == config['PERSON_TWO']:
        return config['PERSON_ONE']
    else:
        raise ValueError("Invalid from address {}".format(from_))


def connect_redis(url):
    host, port = urlparse(url).netloc.split(':')
    return redis.StrictRedis(host=host, port=port)
