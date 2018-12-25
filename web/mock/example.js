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
};
