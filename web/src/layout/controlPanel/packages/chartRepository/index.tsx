import isNull from 'lodash/isNull';
import React, { useContext, useEffect, useState } from 'react';
import { IoMdRefresh, IoMdRefreshCircle } from 'react-icons/io';
import { MdAdd, MdAddCircle } from 'react-icons/md';

import { API } from '../../../../api';
import { AppCtx } from '../../../../context/AppCtx';
import { ChartRepository as ChartRepo } from '../../../../types';
import Loading from '../../../common/Loading';
import NoData from '../../../common/NoData';
import ChartRepositoryCard from './Card';
import styles from './ChartRepository.module.css';
import ChartRepositoryModal from './Modal';

interface ModalStatus {
  open: boolean;
  chartRepository?: ChartRepo;
}

interface Props {
  onAuthError: () => void;
}

const ChartRepository = (props: Props) => {
  const { ctx } = useContext(AppCtx);
  const [isLoading, setIsLoading] = useState(false);
  const [modalStatus, setModalStatus] = useState<ModalStatus>({
    open: false,
  });
  const [chartRepositories, setChartRepositories] = useState<ChartRepo[] | null>(null);
  const selectedOrg = isNull(ctx.org) ? undefined : ctx.org.name;
  const [activeOrg, setActiveOrg] = useState(selectedOrg);

  async function fetchCharts() {
    try {
      setIsLoading(true);
      setChartRepositories(await API.getChartRepositories(activeOrg));
      setIsLoading(false);
    } catch (err) {
      setIsLoading(false);
      if (err.statusText !== 'ErrLoginRedirect') {
        setChartRepositories([]);
      } else {
        props.onAuthError();
      }
    }
  }

  useEffect(() => {
    fetchCharts();
  }, [activeOrg]); /* eslint-disable-line react-hooks/exhaustive-deps */

  useEffect(() => {
    setActiveOrg(selectedOrg);
  }, [selectedOrg]);

  return (
    <>
      <div>
        <div className="d-flex flex-row align-items-center justify-content-between">
          <div className="h3 pb-0">Chart repositories</div>

          <div>
            <button
              className={`btn btn-secondary btn-sm text-uppercase mr-2 ${styles.btnAction}`}
              onClick={fetchCharts}
            >
              <div className="d-flex flex-row align-items-center justify-content-center">
                <IoMdRefresh className="d-inline d-md-none" />
                <IoMdRefreshCircle className="d-none d-md-inline mr-2" />
                <span className="d-none d-md-inline">Refresh</span>
              </div>
            </button>

            <button
              className={`btn btn-secondary btn-sm text-uppercase ${styles.btnAction}`}
              onClick={() => setModalStatus({ open: true })}
            >
              <div className="d-flex flex-row align-items-center justify-content-center">
                <MdAdd className="d-inline d-md-none" />
                <MdAddCircle className="d-none d-md-inline mr-2" />
                <span className="d-none d-md-inline">Add</span>
              </div>
            </button>
          </div>
        </div>
      </div>

      {modalStatus.open && (
        <ChartRepositoryModal
          open={modalStatus.open}
          chartRepository={modalStatus.chartRepository}
          onSuccess={fetchCharts}
          onAuthError={props.onAuthError}
          onClose={() => setModalStatus({ open: false })}
        />
      )}

      {(isLoading || isNull(chartRepositories)) && <Loading />}

      {!isNull(chartRepositories) && (
        <>
          {chartRepositories.length === 0 ? (
            <NoData>
              <>
                <p className="h6 my-4">Add your first chart repository!</p>

                <button type="button" className="btn btn-secondary" onClick={() => setModalStatus({ open: true })}>
                  <div className="d-flex flex-row align-items-center">
                    <MdAddCircle className="mr-2" />
                    <span>Add chart repository</span>
                  </div>
                </button>
              </>
            </NoData>
          ) : (
            <div className="list-group my-4">
              {chartRepositories.map((repo: ChartRepo) => (
                <ChartRepositoryCard
                  key={repo.name}
                  chartRepository={repo}
                  setModalStatus={setModalStatus}
                  onSuccess={fetchCharts}
                  onAuthError={props.onAuthError}
                />
              ))}
            </div>
          )}
        </>
      )}
    </>
  );
};

export default ChartRepository;
