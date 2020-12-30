module.exports = {
  title: 'nimona',
  description: 'a new internet stack, or something like it',
  themeConfig: {
    navbar: false,
    // logo: '/nimona-logo.png',
    search: true,
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Concepts', link: '/concepts/' },
      { text: 'Proposals', link: '/proposals/' },
    ],
    sidebar: [{
      title: 'Nimona',
      path: '/introduction',
      collapsable: false,
      children: [
        '/design-decisions',
      ],
    }, {
      title: 'Concepts',
      collapsable: false,
      children: [
        '/networking',
        {
          title: 'Objects',
          path: '/objects',
          collapsable: false,
          children: [[
            '/proposals/np001-hinted-object-notation',
            'Hinting & Hashing [np001]',
          ], [
            '/proposals/np002-structured-objects',
            'Structure [np002]',
          ]],
          sidebarDepth: 0,
        },
        [
          '/proposals/np003-streams',
          'Streams [np003]',
        ],
        [
          '/proposals/np004-feeds',
          'Feeds [np004]',
        ],
      ],
    }, {
      title: 'Other',
      collapsable: false,
      children: [
        '/proposals/',
      ],
    }],
    displayAllHeaders: true,
  }
}