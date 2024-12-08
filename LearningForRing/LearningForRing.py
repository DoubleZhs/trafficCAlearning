import pandas as pd
import numpy as np
import time
import math
from tqdm import tqdm
import pickle
import matplotlib.pyplot as plt
import seaborn as sns
import random
import xgboost as xgb
from xgboost import plot_importance
from sklearn.model_selection import train_test_split,RandomizedSearchCV
from sklearn.metrics import mean_squared_error,mean_absolute_error, r2_score,mean_absolute_percentage_error


def cut_data(data_df,total_test_df):
    # 根据Date来分
    max_date = data_df['Date'].max()
    train_valid_df = data_df[data_df['Date'].between(max_date-7,max_date-1)].copy()
    test_df = data_df[data_df['Date']==max_date].copy()

    # add statistic Feature  
    od_feature = train_valid_df[train_valid_df['Date'].between(max_date-7,max_date)].groupby(['OD Dig'])['Travel Time'].agg([('od_mean','mean'), ('od_max','max') , ('od_min','min') , ('od_std','std')]).reset_index()
    odh_feature = train_valid_df[train_valid_df['Date'].between(max_date-7,max_date)].groupby(['OD Dig','Hour'])['Travel Time'].agg([('odh_mean','mean'), ('odh_max','max') , ('odh_min','min') , ('odh_std','std')]).reset_index()
    # odq_feature = train_valid_df[train_valid_df['Date'].between(max_date-7,max_date)].groupby(['OD Dig','Quarter'])['Travel Time'].agg([('odq_mean','mean'), ('odq_max','max') , ('odq_min','min') , ('odq_std','std')]).reset_index()

    train_valid_df = pd.merge(train_valid_df,od_feature,on='OD Dig',how='left')
    train_valid_df = pd.merge(train_valid_df,odh_feature,on=['OD Dig','Hour'],how='left')
    # train_valid_df = pd.merge(train_valid_df,odq_feature,on=['OD Dig','Quarter'],how='left')

    test_df = pd.merge(test_df,od_feature,on='OD Dig',how='left')
    test_df = pd.merge(test_df,odh_feature,on=['OD Dig','Hour'],how='left')
    # test_df = pd.merge(test_df,odq_feature,on=['OD Dig','Quarter'],how='left')

    total_test_df = pd.merge(total_test_df,od_feature,on='OD Dig',how='left')
    total_test_df = pd.merge(total_test_df,odh_feature,on=['OD Dig','Hour'],how='left')
    # total_test_df = pd.merge(total_test_df,odq_feature,on=['OD Dig','Quarter'],how='left')
     
    # 最终切分
    train_df,valid_df = train_test_split(train_valid_df, test_size=0.1, random_state=42)

    return train_df, valid_df, test_df, total_test_df, ['od_mean','od_max','od_min','od_std','odh_mean','odh_max','odh_min','odh_std'] # ['od_mean','od_max','od_min','od_std'] # 

def search_params(X_train,y_train,X_val, y_val):
    # 定义 XGBoost 模型
    xgb_model = xgb.XGBRegressor(objective='reg:squarederror',tree_method='hist', enable_categorical=True)

    # 随机搜索参数
    param_dist = {
        'n_estimators': [100, 500, 1000, 2000],
        'learning_rate': [0.01, 0.05, 0.1],
        'max_depth': [3, 5, 6],
        'subsample': [0.6, 0.8, 1.0],
        'colsample_bytree': [0.6, 0.8, 1.0]
    }

    # 随机网格搜索
    random_search = RandomizedSearchCV(estimator=xgb_model, param_distributions=param_dist,
                                       n_iter=50, scoring='neg_mean_absolute_percentage_error', cv=3,
                                       verbose=1, n_jobs=-1, random_state=1024)

    # 执行随机搜索
    random_search.fit(X_train, y_train, 
                      eval_set=[(X_val, y_val)],
                      verbose=False)
    
    # 获取最佳参数
    best_params = random_search.best_params_
    print("Best Parameters:", best_params)

    # 获取最佳模型
    xgb_model = random_search.best_estimator_

    return xgb_model

def getPerformance(y_true , y_pred):
    # 计算回归性能指标
    MSE  = mean_squared_error(y_true,y_pred)
    RMSE = math.sqrt(MSE)
    mae  = mean_absolute_error(y_true,y_pred)
    r2  = r2_score(y_true,y_pred)
    mape = mean_absolute_percentage_error(y_true,y_pred)
    return {'MSE':MSE,'RMSE':RMSE,'MAE':mae,'R2':r2,'MAPE':mape}

def drawImportance(xgb_model):
    # 画出XGBoost模型的重要性
    plt.rcParams['font.sans-serif'] = ['SimHei']
    # print(xgb_model.feature_importances_)
    plot_importance(xgb_model)
    plt.show()

def run(data_df,total_test_df,features,target):
    train_df, valid_df, test_df, total_test_df , add_features = cut_data(data_df,total_test_df)
    new_features = features + add_features

    #print("--------------------------------------------------searching parameters--------------------------------------------------")
    xgb_model = search_params(train_df[new_features],train_df[target],valid_df[new_features],valid_df[target])

    #print("--------------------------------------------------training model--------------------------------------------------")
    xgb_model.set_params(early_stopping_rounds=10,eval_metric='mape')
    xgb_model.fit(train_df[new_features],train_df[target],eval_set=[(valid_df[new_features],valid_df[target])],verbose=False)
    # drawImportance(xgb_model)

    #print("--------------------------------------------------predictions--------------------------------------------------")
    test_df['Predicted Travel Time'] = xgb_model.predict(test_df[new_features])
    total_test_df['Predicted Travel Time'] = xgb_model.predict(total_test_df[new_features])

    performance = getPerformance(test_df[target], test_df['Predicted Travel Time'])
    total_performance = getPerformance(total_test_df[target], total_test_df['Predicted Travel Time'])

    return xgb_model,test_df,total_test_df,performance,total_performance




if __name__ == "__main__":

    for n in [10,20,40,60,80,100,150,200]:

        # 读取数据
        data_df = pd.read_csv(f"Feature/VehicleData_n{n}.csv",encoding='utf-8')


        # 定义特征
        category_features = [] #['Origin','Destination','OD'] 
        digital_features = ['Origin','Destination','O Dig','D Dig',  # 起点与终点特征
                            'Actual In Time','Hour','Quarter', 'Early Commute','Late Commute',  # 进入时间特征
                            'Acceleration','SlowingPro',  # 车辆特征
                            'PathLength','Traffic Light',      # 路径特征            
                            ]
        features = category_features + digital_features
        target = ['Travel Time']

        for cf in category_features:
            data_df[cf] = data_df[cf].astype('category')


        # 日志表
        file_path = f"Learning_Logs/n{n}_learning_logs.txt"
        logfile = open(file_path, "w")
        
        result_list = []

        for Date in tqdm(range(1,21)):
            # 结果表
            performance_list = []
            total_performance_list = []

            sub_data_df = data_df[data_df['Date']<=Date]
            total_test_df = data_df[data_df['Date']==Date]
            logfile.write(f"Date: {Date}" + "\n")

            print(f"-----------------------------------n : {n} ; Date: {Date} ; DiDi Result:-----------------------------------" )
            start_time = time.time()
            # 获取DiDi车
            DiDi_df = sub_data_df[sub_data_df['ClosedVehicle']==True]
            DiDi_model,DiDi_test_df,DiDi_total_test_df,DiDi_performance,DiDi_total_performance = run(DiDi_df,total_test_df,features,target)
            # 存储模型
            with open(f'Model/n{n}_DiDi_vehicle_xgb_model.pkl', 'wb') as file:
                pickle.dump(DiDi_model, file)
            # 增加信息
            DiDi_performance['Type'],DiDi_performance['Note'],DiDi_performance['Pro'],DiDi_performance['n'] = 'DiDi','vehicle',DiDi_df.shape[0]/sub_data_df.shape[0],n
            DiDi_total_performance['Type'],DiDi_total_performance['Note'],DiDi_total_performance['Pro'],DiDi_total_performance['n'] = 'DiDi','vehicle',DiDi_df.shape[0]/sub_data_df.shape[0],n
            
            
            print('pure : MAPE ',DiDi_performance['MAPE'])
            print('total: MAPE ',DiDi_total_performance['MAPE'])
            performance_list.append(DiDi_performance)
            total_performance_list.append(DiDi_total_performance)
            end_time = time.time()
            logfile.write("DiDi Learning Time: {:.2f} seconds".format(end_time - start_time) + "\n")
            DiDi_test_df.to_csv(f"predictResult/n{n}_DiDi_vehicle_predictions.csv",index=None)
            DiDi_total_test_df.to_csv(f"predictResult/n{n}_total_DiDi_vehicle_predictions.csv",index=None)

            # 不断增大比例获得地图车
            for Autonavi_pro in [0.001,0.005,0.01,0.02] + list(np.arange(0,0.31,0.05)[1:]):
                print(f"-----------------------------------n : {n} ; Date: {Date} ; Autonavi {Autonavi_pro} Result:-----------------------------------")
                start_time = time.time()

                Autonavi_pro = round(Autonavi_pro,3)
                Autonavi_df = sub_data_df[(sub_data_df['Tag']<Autonavi_pro)&(sub_data_df['ClosedVehicle']==False)]
                Autonavi_model,Autonavi_test_df,Autonavi_total_test_df,Autonavi_performance,Autonavi_total_performance = run(Autonavi_df,total_test_df,features,target)
                # 存储模型
                with open(f'Model/n{n}_pro{Autonavi_pro}_Autonavi_vehicle_xgb_model.pkl', 'wb') as file:
                    pickle.dump(Autonavi_model, file)
                Autonavi_performance['Type'],Autonavi_performance['Note'],Autonavi_performance['Pro'],Autonavi_performance['n'] = 'Autonavi','vehicle',Autonavi_pro,n
                Autonavi_total_performance['Type'],Autonavi_total_performance['Note'],Autonavi_total_performance['Pro'],Autonavi_total_performance['n'] = 'Autonavi','vehicle',Autonavi_pro,n
                
                print('pure : MAPE ',Autonavi_performance['MAPE'])
                print('total: MAPE ',Autonavi_total_performance['MAPE'])

                performance_list.append(Autonavi_performance)
                total_performance_list.append(Autonavi_total_performance)
                end_time = time.time()
                logfile.write("Autonavi {} Learning Time: {:.2f} seconds".format(Autonavi_pro,end_time - start_time) + "\n")
                Autonavi_test_df.to_csv(f"predictResult/n{n}_Autonavi_Pro{Autonavi_pro}_vehicle_predictions.csv",index=None)
                Autonavi_total_test_df.to_csv(f"predictResult/n{n}_total_Autonavi_Pro{Autonavi_pro}_vehicle_predictions.csv",index=None)
        
            performance_df = pd.DataFrame()
            # 逐行将字典数据添加到 DataFrame
            for row in performance_list:
                temp_df = pd.DataFrame([row])  # 将字典数据转换为DataFrame
                performance_df = pd.concat([performance_df, temp_df], ignore_index=True) 
            performance_df['Date'] = Date
            result_list.append(performance_df)

        result_df = pd.concat(result_list,ignore_index=True)
        result_df.to_csv(f"Performance/n{n}_result.csv",index=None)

        # 关闭文件
        logfile.close()
