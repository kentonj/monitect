module.exports = {
  devServer: {
    proxy: {
      "^/api": {
        target: "http://localhost:8080",
        pathRewrite: { "^/api": "" },
        changeOrigin: true,
        logLevel: "debug"
      }
    }
  }
}
