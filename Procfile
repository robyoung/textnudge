web: gunicorn textnudge.web:app --log-file=-
worker: celery worker --app=textnudge.tasks.app
