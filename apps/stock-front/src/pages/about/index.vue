<template>
  <section class="page">
    <h1>关于</h1>
    <p>本应用位于 <code>apps/vue-project</code>，通过 Host 无界挂载。</p>
    <ul>
      <li>子应用名称（subAppBusName）：<code>vue-project</code></li>
      <li>父→子路由事件：<code>vue-project-route</code></li>
      <li>子→父路由事件：<code>sub-route-change</code></li>
    </ul>
  </section>
</template>

<style scoped>
.page h1 {
  margin: 0 0 12px;
}

.page p,
.page li {
  color: #4b5563;
  line-height: 1.6;
}

code {
  padding: 2px 6px;
  background: #f3f4f6;
  border-radius: 4px;
  font-size: 0.9em;
}
</style>
