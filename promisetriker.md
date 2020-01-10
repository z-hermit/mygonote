```
new Promise((resolve) => {
  setTimeout(() => {
    console.log("timeout")
    resolve()
  }, 2000)
}).then(() => {
  console.log("first then")
  return new Promise(() => {
    console.log("new promise")

  })
})
.then(() => {
  console.log("second then")
})
.then(() => {
  console.log("third then")
})
```
```
Promise.resolve().then(() => {
  console.log("first then")
  return new Promise(() => {
    console.log("new promise")

  })
})
.then(() => {
  console.log("second then")
})
.then(() => {
  console.log("third then")
})
```
结果是一样的，都不会输出second, third then. 原因在于promise处理handler的cb时会放入任务队列, 所以无论已决还是未决， then产生的promise都会是未决， 在到达那个任务队列时变换状态。

resolve(Promise) 或者 then(()=>Promise) 会将新promise替换旧的promise， 意味着旧的promise的defer不会再执行了。