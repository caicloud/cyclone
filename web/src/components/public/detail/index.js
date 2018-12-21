import React from 'react';
import PropTypes from 'prop-types';
import Head from './Head';
import HeadItem from './HeadItem';
import DetailContent from './Content';

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
