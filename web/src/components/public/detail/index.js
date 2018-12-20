import React from 'react';
import PropTypes from 'prop-types';
import Head from './head';
import HeadItem from './headItem';
import DetailContent from './content';

class Detail extends React.Component {
  render() {
    const { children } = this.props;
    return <div className="u-detail">{children}</div>;
  }
}

Detail.DetailHead = Head;
Detail.DetailHeadItem = HeadItem;
Detail.DetailContent = DetailContent;

Detail.propTypes = {
  children: PropTypes.any,
};
export default Detail;
