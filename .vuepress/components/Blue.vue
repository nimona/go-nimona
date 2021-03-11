<template>
  <section class="welcome">
    <script async defer data-domain="nimona.io" src="https://plausible.io/js/plausible.js"></script>
    <div class="container">
      <h1 class="welcome__title">nimona</h1>
      <p class="welcome__description">
        a new internet stack; or something like it
        <br>
        <small><a href="/docs/">check out the docs</a></small>,
        <small><a href="https://github.com/nimona/go-nimona">contribute on github</a></small>
      </p>
      <figure class="welcome__image">
        <img src="/nimona-logo-icon-white.png" />
      </figure>
    </div>
  </section>
</template>

<script>
import NavLink from '@theme/components/NavLink.vue'

export default {
  name: 'Home',

  components: { NavLink },

  computed: {
    data () {
      return this.$page.frontmatter
    },

    actionLink () {
      return { }
    }
  }
}
</script>

<style>
  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    -webkit-tap-highlight-color: rgba(0,0,0,0);
  }
  body {
    position: relative;
    min-height: 100vh;
    font: 16px Helvetica,sans-serif;
  }
  .welcome {
    position: relative;
    z-index: 0;
    background: #6697FF;
    background: linear-gradient(45deg,#6697FF,#95C6FF);
  }
  .welcome a {
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    outline: none;
    font-size: inherit;
    color: inherit;
    text-decoration: inherit;
    text-transform: inherit;
    background-color: transparent;
    border-bottom: 1px dotted #ffffff9e;
    display: inline-block;
    line-height: 0.85;
  }
  .container {
    width: 100%;
    max-width: 640px;
    margin-left: auto;
    margin-right: auto;
    padding-left: 24px;
    padding-right: 24px;
  }
  .welcome .container {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    justify-content: center;
    height: 100vh;
    min-height: 450px;
    color: #fff;
  }
  .welcome__image {
    display: none
  }
  @media (min-width: 640px) {
    .container {
      padding-left: 48px;
      padding-right: 48px;
    }
    .welcome__title {
      font-size: 3.6em;
    }
    .welcome__description {
      width: 80%;
    }
    .welcome__image {
      display: flex;
      justify-content: center;
      align-items: center;
      z-index: -1;
      position: absolute;
      top: 50%;
      right: 0;
      height: 70%;
      width: 24%;
      overflow: hidden;
      -webkit-transform: translateY(-50%);
      transform: translateY(-50%);
    }
    .welcome__image img {
      display: block;
      margin: auto;
      max-width: 150px;
      max-height: 100%
    }
  }
  .header {
    position: absolute;
    z-index: 10;
    top: 12px;
    left: 0;
    width: 100%;
  }
  .header .container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    font-size: .8em;
    font-weight: 700;
    text-transform: uppercase;
  }
  .header__menu {
    display: flex;
    color: #fff;
    list-style: none;
  }
  .header__menu-item {
    margin-right: 8px;
  }
  .header__menu-item__label {
    display: block;
    width: 100%;
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    border: none;
    outline: none;
    font-size: inherit;
    color: inherit;
    text-decoration: inherit;
    text-transform: inherit;
    text-align: center;
    padding: 12px 16px;
    background-color: transparent;
    border-radius: 4px;
    font-weight: 400;
  }
  .welcome__actions {
    display: flex;
    margin-top: 40px;
    flex-wrap: wrap;
  }
  .welcome__action {
    padding: 12px 16px;
    border-radius: 50px;
    text-decoration: none;
    color: #fff;
    border: 1px solid #fff;
    margin-right: 24px;
    margin-bottom: 24px;
    transition: background-color .3s ease,color .3s ease;
    min-width: -webkit-fit-content;
    min-width: -moz-fit-content;
    min-width: fit-content;
  }
  .welcome__action:hover {
    background-color: #fff;
    color: #666;
  }
</style>