module.exports = {
  // return data directly
  'GET /api/example': {
    hello: 'world',
  },

  'GET /api/example/func': (req, res) => {
    res.json({
      query: req.query,
    });
  },

  'GET /api/v1/integrations': (req, res) => {
    res.json([
      {
        metadata: { name: 'sonar1', description: '1111' },
        spec: {
          type: 'SonarQube',
          sonarqube: {
            server: 'http://192.168.21.100:9000',
            token: '11111',
            creationTime: '2018-01-01',
          },
        },
      },
      {
        metadata: { name: 'scm1', description: '1111' },
        spec: {
          type: 'SCM',
          scm: {
            type: 'Github',
            server: 'http://192.168.21.100:9000',
            username: '11111',
          },
        },
      },
      {
        metadata: { name: 'scm2', description: '1111' },
        spec: {
          type: 'SCM',
          scm: {
            type: 'Gitlab',
            server: 'http://192.168.21.100:9000',
            username: '2222',
          },
        },
      },
      {
        metadata: { name: 'docker1', description: '1111' },
        spec: {
          type: 'DockerRegistry',
          dockerRegistry: {
            server: 'http://192.168.21.100:9000',
            username: '2222',
          },
        },
      },
    ]);
  },
};
