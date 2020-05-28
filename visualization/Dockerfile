FROM node:10.19

WORKDIR /etc/app

EXPOSE 3000

RUN npm install -g react-scripts@3.4.1
RUN npm install

COPY entrypoint.sh .

CMD [ "./entrypoint.sh" ]
