import { Table, Collapse } from 'antd';
import PropTypes from 'prop-types';

const Inputs = ({ inputs = {} }) => {
  const resourceColumns = [
    {
      title: intl.get('name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.get('type'),
      dataIndex: 'type',
    },
    {
      title: intl.get('path'),
      dataIndex: 'path',
    },
  ];

  const argumentColumns = [
    {
      title: intl.get('name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.get('value'),
      dataIndex: 'value',
      key: 'value',
    },
  ];

  return (
    <Collapse activeKey={['1', '2']}>
      <Collapse.Panel showArrow={false} header={intl.get('resources')} key="1">
        <Table
          columns={resourceColumns}
          dataSource={inputs.resources}
          pagination={false}
          rowKey="name"
        />
      </Collapse.Panel>
      <Collapse.Panel
        showArrow={false}
        header={intl.get('stage.input.arguments')}
        key="2"
      >
        <Table
          columns={argumentColumns}
          dataSource={inputs.arguments}
          pagination={false}
          rowKey="name"
        />
      </Collapse.Panel>
    </Collapse>
  );
};

Inputs.propTypes = {
  inputs: PropTypes.shape({
    resources: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string,
        path: PropTypes.string,
        type: PropTypes.string,
      })
    ).isRequired,
    arguments: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string,
        value: PropTypes.string,
      })
    ).isRequired,
  }),
};

export default Inputs;
