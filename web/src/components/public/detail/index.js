import React from 'react';
import PropTypes from 'prop-types';
import Head from './Head';
import HeadItem from './HeadItem';
import DetailContent from './Content';
import Action from './Action';

class Detail extends React.Component {
  render() {
    const { children, actions } = this.props;
    return (
      <div className="u-detail">
        {actions && actions}
        {children}
      </div>
    );
  }
}

Detail.DetailAction = Action;
Detail.DetailHead = Head;
Detail.DetailHeadItem = HeadItem;
Detail.DetailContent = DetailContent;

Detail.propTypes = {
  children: PropTypes.any,
  actions: PropTypes.node,
};
export default Detail;
