module.exports = {
  'GET /apis/v1alpha1/stages/templates': [
    {
      metadata: {
        name: 'ImageBuild',
        description: 'Build an image from a Dockerfile',
        tag: ['build'],
      },
    },
    {
      metadata: {
        name: 'ImagePublish',
        description: 'Publish the image to docker registry',
        tag: ['publish'],
      },
    },
    {
      metadata: {
        name: 'MavenTest',
        description: 'Run test by Maven',
        tag: ['test'],
      },
    },
  ],
};
