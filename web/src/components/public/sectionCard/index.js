import styles from './index.module.less';
import PropTypes from 'prop-types';

class SectionCard extends React.Component {
  render() {
    const { title, children } = this.props;
    return (
      <div className={styles['u-section-card']}>
        <div className={styles['title']}>{title}</div>
        <div className={styles['content']}>{children}</div>
      </div>
    );
  }
}

SectionCard.propTypes = {
  title: PropTypes.string,
  children: PropTypes.any,
};
export default SectionCard;
