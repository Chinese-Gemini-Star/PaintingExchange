# 使用 Nginx 官方镜像
FROM nginx:alpine

# 创建目录
RUN mkdir -p /usr/share/nginx/html/front /usr/share/nginx/html/back

# 复制前端项目到容器
COPY front/ /usr/share/nginx/html/front
COPY back/ /usr/share/nginx/html/back

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]