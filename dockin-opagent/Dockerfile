FROM centos:7

WORKDIR /data/app
ENV DOCKER  true

ADD release/*.tar.gz /data/app/

ENTRYPOINT ["sh", "-c", "/data/app/dockin-opagent/bin/start.sh"]
