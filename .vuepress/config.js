module.exports = {
  title: 'nimona',
  description: 'a new internet stack, or something like it',
  head: [[
    'meta', {
      'name': 'go-import',
      'content': 'nimona.io git https://github.com/nimona/go-nimona',
    }, ''
  ]],
  themeConfig: {
    navbar: false,
    // logo: '/docs/nimona-logo.png',
    search: true,
    nav: [
      { text: 'Home', link: '/docs/' },
      { text: 'Concepts', link: '/docs/concepts/' },
      { text: 'Proposals', link: '/docs/proposals/' },
    ],
    sidebar: [{
      title: 'Documentation',
      collapsable: false,
      children: [
        '/docs/',
        '/docs/design-decisions',
      ],
      sidebarDepth: 0,
    }, {
      title: 'Concepts',
      collapsable: false,
      children: [
        '/docs/networking',
        {
          title: 'Objects',
          path: '/docs/objects',
          collapsable: false,
          children: [[
            '/docs/proposals/np001-hinted-object-notation',
            'Hinting & Hashing [np001]',
          ], [
            '/docs/proposals/np002-structured-objects',
            'Structure [np002]',
          ]],
          sidebarDepth: 0,
        },
        [
          '/docs/proposals/np003-streams',
          'Streams [np003]',
        ],
        [
          '/docs/proposals/np004-feeds',
          'Feeds [np004]',
        ],
      ],
    }, {
      title: 'Other',
      collapsable: false,
      children: [
        '/docs/proposals/',
      ],
    }],
    displayAllHeaders: true,
  },
  plugins: [[
    'vuepress-plugin-clean-urls',
    {
      normalSuffix: '/',
      indexSuffix: '/',
      notFoundPath: '/404.html',
    },
  ]],
}