import KeyValue from '@/components/public/KeyValue';
import { Table } from 'antd';

import style from './detail.module.less';

const Configuration = ({ configuration = {} }) => {
  const containers = _.get(configuration, 'containers', []);
  const envColumns = [
    {
      title: intl.get('stage.spec.container.env.name'),
      dataIndex: 'name',
      width: '160px',
    },
    {
      title: intl.get('stage.spec.container.env.value'),
      dataIndex: 'value',
      width: '160px',
    },
  ];
  return _.map(containers, (container, index) => (
    <div key={container.name || index}>
      <KeyValue
        cls={style['kv-item']}
        name={intl.get('stage.spec.container.image')}
        value={container.image}
      />
      <KeyValue
        cls={style['kv-item']}
        name={'CMD'}
        isEmpty={_.isEmpty(container.args)}
        value={_.last(container.args)}
      />
      <KeyValue
        cls={style['kv-item']}
        name={intl.get('stage.spec.container.envs')}
        isEmpty={_.isEmpty(container.env)}
        value={
          <Table
            columns={envColumns}
            dataSource={container.env}
            size="small"
            pagination={false}
          />
        }
      />
    </div>
  ));
};

export default Configuration;
