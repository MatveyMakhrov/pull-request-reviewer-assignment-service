FROM grafana/k6:latest

WORKDIR /loadtest

COPY loadtest.js .

CMD ["run", "loadtest.js"]