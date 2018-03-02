FROM trestletech/plumber

RUN R -e "install.packages('e1071')"
RUN R -e "install.packages('pracma')"
RUN R -e "install.packages('plumber')"

ADD api.R /qeeg-api/api.R

EXPOSE 80
ENTRYPOINT ["R", "-e", "pr <- plumber::plumb(commandArgs()[4]); pr$run(host='https://api.fredhw.me', port=80)"]
CMD ["/qeeg-api/api.R"]