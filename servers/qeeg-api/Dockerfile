FROM trestletech/plumber

RUN R -e "install.packages('e1071')"
RUN R -e "install.packages('pracma')"

COPY . /app/

EXPOSE 80

ENTRYPOINT ["R", "-e", "setwd('/app/'); pr <- plumber::plumb(commandArgs()[4]); pr$run(host='0.0.0.0', port=80)"]
CMD ["/app/api.R"]