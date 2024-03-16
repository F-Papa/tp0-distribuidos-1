FROM python:3.9.7-slim
RUN apt-get update
RUN apt-get install -y netcat
COPY test_script /
ENTRYPOINT ["python", "/main.py"]
# ENTRYPOINT ["/bin/bash"]