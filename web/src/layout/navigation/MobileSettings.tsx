import classnames from 'classnames';
import isNull from 'lodash/isNull';
import isUndefined from 'lodash/isUndefined';
import React, { useContext, useState } from 'react';
import { FaUserCircle } from 'react-icons/fa';
import { GoThreeBars } from 'react-icons/go';

import { AppCtx } from '../../context/AppCtx';
import Sidebar from '../common/Sidebar';
import LogOut from './LogOut';
import styles from './MobileSettings.module.css';

interface Props {
  setOpenSignUp: React.Dispatch<React.SetStateAction<boolean>>;
  setOpenLogIn: React.Dispatch<React.SetStateAction<boolean>>;
  privateRoute?: boolean;
}

const MobileSettings = (props: Props) => {
  const { ctx } = useContext(AppCtx);
  const [openSideBarStatus, setOpenSideBarStatus] = useState(false);

  return (
    <div className={`btn-group navbar-toggler pr-0 ${styles.navbarToggler}`}>
      {isUndefined(ctx.user) ? (
        <div className="spinner-grow spinner-grow-sm text-light" role="status">
          <span className="sr-only">Loading...</span>
        </div>
      ) : (
        <Sidebar
          className="d-inline-block d-md-none"
          buttonType="position-relative btn text-secondary pr-0 pl-3"
          buttonIcon={
            <div
              className={classnames(
                'rounded-circle d-flex align-items-center justify-content-center',
                styles.iconWrapper
              )}
            >
              {!isNull(ctx.user) ? <FaUserCircle /> : <GoThreeBars />}
            </div>
          }
          direction="right"
          header={
            <>
              {!isNull(ctx.user) && (
                <div className="h6 mb-0">
                  Signed in as <span className="font-weight-bold">{ctx.user.alias}</span>
                </div>
              )}
            </>
          }
          open={openSideBarStatus}
          onOpenStatusChange={(status: boolean) => setOpenSideBarStatus(status)}
        >
          <>
            {!isUndefined(ctx.user) && (
              <>
                {!isNull(ctx.user) ? (
                  <>
                    {/* TODO - Control panel mobile version */}
                    {/* <Link
                      className="dropdown-item my-2"
                      to={{
                        pathname: '/control-panel',
                      }}
                      onClick={() => setOpenSideBarStatus(false)}
                    >
                      Control Panel
                    </Link> */}

                    <LogOut
                      className="my-2"
                      onSuccess={() => setOpenSideBarStatus(false)}
                      privateRoute={props.privateRoute}
                    />
                  </>
                ) : (
                  <>
                    <button
                      className="dropdown-item my-2"
                      onClick={() => {
                        setOpenSideBarStatus(false);
                        props.setOpenLogIn(true);
                      }}
                    >
                      Sign in
                    </button>

                    <button
                      className="dropdown-item my-2"
                      onClick={() => {
                        setOpenSideBarStatus(false);
                        props.setOpenSignUp(true);
                      }}
                    >
                      Sign up
                    </button>
                  </>
                )}
              </>
            )}
          </>
        </Sidebar>
      )}
    </div>
  );
};

export default MobileSettings;
