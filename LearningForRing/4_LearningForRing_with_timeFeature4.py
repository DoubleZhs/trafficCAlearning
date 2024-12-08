import os
import datetime
import csv
import pickle
import math
import time
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import xgboost as xgb
from tqdm import tqdm
from xgboost import plot_importance
from sklearn.model_selection import train_test_split,RandomizedSearchCV
from sklearn.metrics import mean_squared_error,mean_absolute_error, r2_score,mean_absolute_percentage_error


is_search = False

# 文件note
note = '天数无上限&测试集另取3天'

if not os.path.exists(f'Model/{note}'):
    os.makedirs(f'Model/{note}')

if not os.path.exists(f'TimeFeature_predictResult/{note}'):
    os.makedirs(f'TimeFeature_predictResult/{note}')


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
                                       n_iter=30, scoring='neg_mean_absolute_percentage_error', cv=3,
                                       verbose=1, n_jobs=-1, random_state=42)

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


def run(train_df, valid_df, test_df, total_test_df,features,target):

    if is_search:
        #print("--------------------------------------------------searching parameters--------------------------------------------------")
        xgb_model = search_params(train_df[features],train_df[target],valid_df[features],valid_df[target])

        #print("--------------------------------------------------training model--------------------------------------------------")
        xgb_model.set_params(early_stopping_rounds=10,eval_metric='mape')
    else:
        #print("--------------------------------------------------training model--------------------------------------------------")
        xgb_model = xgb.XGBRegressor(objective='reg:squarederror',tree_method='hist', enable_categorical=True)
        xgb_model.set_params(n_estimators=2000, 
                             learning_rate=0.05, 
                             max_depth=6, 
                             subsample=1, 
                             colsample_bytree=0.8, 
                             early_stopping_rounds=10, 
                             eval_metric='mape')
    
    xgb_model.fit(train_df[features],train_df[target],eval_set=[(valid_df[features],valid_df[target])],verbose=False)
    # drawImportance(xgb_model)

    #print("--------------------------------------------------predictions--------------------------------------------------")
    test_df['Predicted Travel Time'] = xgb_model.predict(test_df[features])
    total_test_df['Predicted Travel Time'] = xgb_model.predict(total_test_df[features])

    performance = getPerformance(test_df[target], test_df['Predicted Travel Time'])
    total_performance = getPerformance(total_test_df[target], total_test_df['Predicted Travel Time'])

    return xgb_model,test_df,total_test_df,performance,total_performance










if __name__ == '__main__':
    # 加载数据
    n_lst = [10,20,40,60,80,100,150,200]
    pro_lst = [0.001,0.005,0.01] + list(np.arange(0,0.31,0.05)[1:])
    date_lst = range(1,29)

    # 设定特征
    features = ['Origin','Destination','O Dig','D Dig',  # 起点与终点特征
                'Actual In Time','Hour','Quarter',   # 进入时间特征
                'Acceleration','SlowingPro',  # 车辆特征
                'PathLength','Traffic Light', # 路径特征            
                ] + [f'mean_TravelTime_before_{i}' for i in range(1,7)] + [f'std_TravelTime_before_{i}' for i in range(1,7)] # 时序特征
    target = ['Travel Time']

    # 新建日志
    log_path = f"Learning_Logs/learning_logs_timeFeature_{note}.txt"
    with open(log_path, 'w') as file:
        pass

    # 新建performance
    performance_path = f"Performance/performance_timeFeature_{note}.csv"
    fieldnames = ['MSE','RMSE','MAE','R2','MAPE','Type','Note','Pro','n','date']
    with open(performance_path, 'w') as file:
        writer = csv.DictWriter(file, fieldnames=fieldnames)
        # 写入表头
        writer.writeheader()




    for n in n_lst:
        
        # 读取测试数据
        DIDI_TEST_DF = pd.read_csv(f"TimeFeature/test/VehicleData_n{n}.csv",encoding='utf-8')
        TOTAL_TEST_DF = pd.read_csv(f"TimeFeature/test/VehicleData_n{n}_total.csv",encoding='utf-8')

        # 读取训练数据
        DiDi_df = pd.read_csv(f"TimeFeature/train/VehicleData_n{n}.csv",encoding='utf-8')
        data_df = pd.read_csv(f"TimeFeature/train/VehicleData_n{n}_total.csv",encoding='utf-8')


        for date in date_lst:

            if date < 10:
                test_date1,test_date2 = 2,4
            elif date < 20:
                test_date1,test_date2 = 9,11
            else:
                test_date1,test_date2 = 16,18
            
            print(f"-----------------------------------n : {n} ; Date: {date} ; DiDi Result:-----------------------------------")
            with open(log_path, 'a') as file:
                file.write(f"-----------------------------------n : {n} ; Date: {date} ; DiDi Result:-----------------------------------\n")

            start_time = time.time()

            # DiDi训练与验证数据
            DiDi_train_valid_df = DiDi_df[(DiDi_df['Date']<=date)]
            DiDi_train_df,DiDi_valid_df = train_test_split(DiDi_train_valid_df, test_size=0.2, random_state=42,shuffle=True)
            
            # DiDi测试数据
            DiDi_test_df  = DIDI_TEST_DF[(DIDI_TEST_DF['Date'].between(test_date1,test_date2))].reset_index(drop=True)

            # DiDi总的测试数据
            DiDi_total_test_df = TOTAL_TEST_DF[TOTAL_TEST_DF['Date'].between(test_date1,test_date2)].reset_index(drop=True)


            DiDi_model,DiDi_test_df,DiDi_total_test_df,DiDi_performance,DiDi_total_performance = run(DiDi_train_df, DiDi_valid_df, DiDi_test_df, DiDi_total_test_df,features,target)

            with open(f'Model/{note}/TimeFeature_DiDi_n{n}_date{date}_xgb_model.pkl', 'wb') as file:
                pickle.dump(DiDi_model, file)


            DiDi_performance['Type'],DiDi_performance['Note'],DiDi_performance['Pro'],DiDi_performance['n'],DiDi_performance['date'] = 'DiDi','pure',data_df[(data_df['Date']<=date)&(data_df['ClosedVehicle']==True)].shape[0]/data_df[data_df['Date']<=date].shape[0],n,date
            DiDi_total_performance['Type'],DiDi_total_performance['Note'],DiDi_total_performance['Pro'],DiDi_total_performance['n'],DiDi_total_performance['date'] = 'DiDi','total',data_df[(data_df['Date']<=date)&(data_df['ClosedVehicle']==True)].shape[0]/data_df[data_df['Date']<=date].shape[0],n,date
            
            end_time = time.time()
            

            print('pure : MAPE '+str(DiDi_performance['MAPE']))
            print('total: MAPE '+str(DiDi_total_performance['MAPE']))
            print(f"Time : {end_time - start_time} seconds")
            with open(log_path, 'a') as file:
                file.write('pure : MAPE '+str(DiDi_performance['MAPE'])+'\n')
                file.write('total: MAPE '+str(DiDi_total_performance['MAPE'])+'\n')
                file.write(f"Time : {end_time - start_time} seconds"+'\n')

            with open(performance_path, 'a') as file:
                writer = csv.DictWriter(file, fieldnames=fieldnames)
                # 逐行写入数据
                writer.writerow(DiDi_performance)
                writer.writerow(DiDi_total_performance)

            DiDi_test_df.to_csv(f"TimeFeature_predictResult/{note}/TimeFeature_DiDi_predictions_n{n}_date{date}.csv",index=None)
            # DiDi_total_test_df.to_csv(f"TimeFeature_predictResult/{note}/TimeFeature_total_DiDi_predictions_n{n}_date{date}.csv",index=None)

            # 高德数据
            for Autonavi_pro in pro_lst:
                Autonavi_pro = round(Autonavi_pro,3)
                
                # 高德
                AUTONAVI_TEST_DF = pd.read_csv(f"TimeFeature/test/VehicleData_n{n}_pro{Autonavi_pro}.csv",encoding='utf-8')
    
                Autonavi_df = pd.read_csv(f"TimeFeature/train/VehicleData_n{n}_pro{Autonavi_pro}.csv",encoding='utf-8')

                Autonavi_total_test_df = TOTAL_TEST_DF[TOTAL_TEST_DF['Date'].between(test_date1,test_date2)].reset_index(drop=True)


                print(f"-----------------------------------n : {n} ; Date: {date} ; Autonavi {Autonavi_pro} Result:-----------------------------------")
                with open(log_path, 'a') as file:
                    file.write(f"-----------------------------------n : {n} ; Date: {date} ; Autonavi {Autonavi_pro} Result:-----------------------------------\n")

                start_time = time.time()

                Autonavi_train_valid_df = Autonavi_df[(Autonavi_df['Date']<=date)].reset_index(drop=True)
                Autonavi_train_df,Autonavi_valid_df = train_test_split(Autonavi_train_valid_df, test_size=0.2, random_state=42,shuffle=True)
                
                Autonavi_test_df  = AUTONAVI_TEST_DF[(AUTONAVI_TEST_DF['Date'].between(test_date1,test_date2))].reset_index(drop=True)

                Autonavi_model,Autonavi_test_df,Autonavi_total_test_df,Autonavi_performance,Autonavi_total_performance = run(Autonavi_train_df,Autonavi_valid_df,Autonavi_test_df,Autonavi_total_test_df,features,target)

                with open(f'Model/{note}/TimeFeature_DiDi_n{n}_Autonavi_pro{Autonavi_pro}_date{date}_xgb_model.pkl', 'wb') as file:
                    pickle.dump(Autonavi_model, file)

                Autonavi_performance['Type'],Autonavi_performance['Note'],Autonavi_performance['Pro'],Autonavi_performance['n'],Autonavi_performance['date'] = 'Autonavi','pure',Autonavi_pro,n,date
                Autonavi_total_performance['Type'],Autonavi_total_performance['Note'],Autonavi_total_performance['Pro'],Autonavi_total_performance['n'],Autonavi_total_performance['date'] = 'Autonavi','total',Autonavi_pro,n,date

                end_time = time.time()
                

                print('pure : MAPE '+str(Autonavi_performance['MAPE']))
                print('total: MAPE '+str(Autonavi_total_performance['MAPE']))
                print(f"Time : {end_time - start_time} seconds")
                with open(log_path, 'a') as file:
                    file.write('pure : MAPE '+str(Autonavi_performance['MAPE'])+'\n')
                    file.write('total: MAPE '+str(Autonavi_total_performance['MAPE'])+'\n')
                    file.write(f"Time : {end_time - start_time} seconds"+'\n')

                with open(performance_path, 'a') as file:
                    writer = csv.DictWriter(file, fieldnames=fieldnames)
                    # 逐行写入数据
                    writer.writerow(Autonavi_performance)
                    writer.writerow(Autonavi_total_performance)


                Autonavi_test_df.to_csv(f"TimeFeature_predictResult/{note}/TimeFeature_Autonavi_predictions_n{n}_Pro{Autonavi_pro}_date{date}.csv",index=None)
                # Autonavi_total_test_df.to_csv(f"TimeFeature_predictResult/{note}/TimeFeature_total_Autonavi_predictions_n{n}_Pro{Autonavi_pro}_date{date}.csv",index=None)

                