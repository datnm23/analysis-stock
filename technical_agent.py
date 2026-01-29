"""
Technical Analysis Agent for Vietnamese Stock Market
Phân tích kỹ thuật chứng khoán Việt Nam
"""

import pandas as pd
import numpy as np
import ta
from typing import Dict, List, Optional
from datetime import datetime, timedelta
import vnstock
from vnstock import stock_historical_data


class TechnicalAnalysisAgent:
    """
    Agent chuyên phân tích kỹ thuật cho thị trường chứng khoán VN
    """
    
    def __init__(self, symbol: str, days: int = 90):
        """
        Initialize agent with stock symbol
        
        Args:
            symbol: Mã chứng khoán (VD: 'VNM', 'HPG')
            days: Số ngày lịch sử để phân tích
        """
        self.symbol = symbol.upper()
        self.days = days
        self.data = None
        self.indicators = {}
        
    def fetch_data(self) -> pd.DataFrame:
        """
        Lấy dữ liệu lịch sử giá từ vnstock
        """
        try:
            end_date = datetime.now()
            start_date = end_date - timedelta(days=self.days)
            
            # Fetch từ vnstock
            self.data = stock_historical_data(
                symbol=self.symbol,
                start_date=start_date.strftime('%Y-%m-%d'),
                end_date=end_date.strftime('%Y-%m-%d'),
                resolution='1D'
            )
            
            # Chuẩn hóa column names
            self.data.columns = [col.lower() for col in self.data.columns]
            
            # Đảm bảo có đủ các cột cần thiết
            required_cols = ['open', 'high', 'low', 'close', 'volume']
            if not all(col in self.data.columns for col in required_cols):
                raise ValueError(f"Missing required columns in data")
            
            return self.data
            
        except Exception as e:
            print(f"Error fetching data for {self.symbol}: {e}")
            raise
    
    def calculate_all_indicators(self) -> Dict:
        """
        Tính toán tất cả các chỉ báo kỹ thuật
        """
        if self.data is None:
            self.fetch_data()
        
        df = self.data.copy()
        
        # 1. RSI (Relative Strength Index)
        self.indicators['rsi'] = ta.momentum.RSIIndicator(
            close=df['close'], 
            window=14
        ).rsi()
        
        # 2. MACD (Moving Average Convergence Divergence)
        macd = ta.trend.MACD(
            close=df['close'],
            window_slow=26,
            window_fast=12,
            window_sign=9
        )
        self.indicators['macd'] = macd.macd()
        self.indicators['macd_signal'] = macd.macd_signal()
        self.indicators['macd_diff'] = macd.macd_diff()
        
        # 3. Bollinger Bands
        bb = ta.volatility.BollingerBands(
            close=df['close'],
            window=20,
            window_dev=2
        )
        self.indicators['bb_upper'] = bb.bollinger_hband()
        self.indicators['bb_middle'] = bb.bollinger_mavg()
        self.indicators['bb_lower'] = bb.bollinger_lband()
        self.indicators['bb_width'] = bb.bollinger_wband()
        
        # 4. Moving Averages
        self.indicators['sma_20'] = ta.trend.SMAIndicator(
            close=df['close'], 
            window=20
        ).sma_indicator()
        
        self.indicators['sma_50'] = ta.trend.SMAIndicator(
            close=df['close'],
            window=50
        ).sma_indicator()
        
        self.indicators['ema_12'] = ta.trend.EMAIndicator(
            close=df['close'],
            window=12
        ).ema_indicator()
        
        self.indicators['ema_26'] = ta.trend.EMAIndicator(
            close=df['close'],
            window=26
        ).ema_indicator()
        
        # 5. Stochastic Oscillator
        stoch = ta.momentum.StochasticOscillator(
            high=df['high'],
            low=df['low'],
            close=df['close'],
            window=14,
            smooth_window=3
        )
        self.indicators['stoch_k'] = stoch.stoch()
        self.indicators['stoch_d'] = stoch.stoch_signal()
        
        # 6. ADX (Average Directional Index)
        adx = ta.trend.ADXIndicator(
            high=df['high'],
            low=df['low'],
            close=df['close'],
            window=14
        )
        self.indicators['adx'] = adx.adx()
        self.indicators['adx_pos'] = adx.adx_pos()
        self.indicators['adx_neg'] = adx.adx_neg()
        
        # 7. Volume Indicators
        self.indicators['volume_sma'] = ta.volume.VolumeWeightedAveragePrice(
            high=df['high'],
            low=df['low'],
            close=df['close'],
            volume=df['volume']
        ).volume_weighted_average_price()
        
        # 8. ATR (Average True Range) - Volatility
        self.indicators['atr'] = ta.volatility.AverageTrueRange(
            high=df['high'],
            low=df['low'],
            close=df['close'],
            window=14
        ).average_true_range()
        
        return self.indicators
    
    def generate_signals(self) -> Dict:
        """
        Tạo tín hiệu giao dịch dựa trên các chỉ báo
        """
        if not self.indicators:
            self.calculate_all_indicators()
        
        signals = {
            'symbol': self.symbol,
            'timestamp': datetime.now().isoformat(),
            'current_price': float(self.data['close'].iloc[-1]),
            'recommendation': 'HOLD',
            'confidence': 0.5,
            'signals': [],
            'reasons': [],
            'technical_score': 0
        }
        
        score = 0
        
        # === RSI Analysis ===
        current_rsi = self.indicators['rsi'].iloc[-1]
        if not np.isnan(current_rsi):
            if current_rsi < 30:
                score += 2
                signals['signals'].append('RSI_OVERSOLD')
                signals['reasons'].append(f'RSI quá bán ({current_rsi:.1f} < 30) - Tín hiệu mua mạnh')
            elif current_rsi < 40:
                score += 1
                signals['signals'].append('RSI_LOW')
                signals['reasons'].append(f'RSI thấp ({current_rsi:.1f}) - Xu hướng tăng có thể')
            elif current_rsi > 70:
                score -= 2
                signals['signals'].append('RSI_OVERBOUGHT')
                signals['reasons'].append(f'RSI quá mua ({current_rsi:.1f} > 70) - Nguy cơ điều chỉnh')
            elif current_rsi > 60:
                score -= 1
                signals['signals'].append('RSI_HIGH')
                signals['reasons'].append(f'RSI cao ({current_rsi:.1f}) - Cần thận trọng')
        
        # === MACD Analysis ===
        current_macd = self.indicators['macd'].iloc[-1]
        current_macd_signal = self.indicators['macd_signal'].iloc[-1]
        prev_macd = self.indicators['macd'].iloc[-2]
        prev_macd_signal = self.indicators['macd_signal'].iloc[-2]
        
        if not any(np.isnan([current_macd, current_macd_signal, prev_macd, prev_macd_signal])):
            # MACD crossover
            if prev_macd <= prev_macd_signal and current_macd > current_macd_signal:
                score += 2
                signals['signals'].append('MACD_BULLISH_CROSS')
                signals['reasons'].append('MACD cắt lên Signal - Tín hiệu tăng')
            elif prev_macd >= prev_macd_signal and current_macd < current_macd_signal:
                score -= 2
                signals['signals'].append('MACD_BEARISH_CROSS')
                signals['reasons'].append('MACD cắt xuống Signal - Tín hiệu giảm')
            
            # MACD position
            if current_macd > 0:
                score += 0.5
            else:
                score -= 0.5
        
        # === Moving Average Analysis ===
        current_price = self.data['close'].iloc[-1]
        sma_20 = self.indicators['sma_20'].iloc[-1]
        sma_50 = self.indicators['sma_50'].iloc[-1]
        ema_12 = self.indicators['ema_12'].iloc[-1]
        
        if not np.isnan(sma_20):
            if current_price > sma_20:
                score += 1
                signals['signals'].append('PRICE_ABOVE_SMA20')
                signals['reasons'].append(f'Giá trên SMA20 ({sma_20:.1f}) - Xu hướng tăng ngắn hạn')
            else:
                score -= 1
                signals['signals'].append('PRICE_BELOW_SMA20')
                signals['reasons'].append(f'Giá dưới SMA20 ({sma_20:.1f}) - Xu hướng giảm ngắn hạn')
        
        if not any(np.isnan([sma_20, sma_50])):
            if sma_20 > sma_50:
                score += 1
                signals['signals'].append('GOLDEN_CROSS')
                signals['reasons'].append('SMA20 > SMA50 - Golden Cross, xu hướng tăng')
            else:
                score -= 1
                signals['signals'].append('DEATH_CROSS')
                signals['reasons'].append('SMA20 < SMA50 - Death Cross, xu hướng giảm')
        
        # === Bollinger Bands Analysis ===
        bb_upper = self.indicators['bb_upper'].iloc[-1]
        bb_lower = self.indicators['bb_lower'].iloc[-1]
        bb_middle = self.indicators['bb_middle'].iloc[-1]
        
        if not any(np.isnan([bb_upper, bb_lower, bb_middle])):
            if current_price < bb_lower:
                score += 1.5
                signals['signals'].append('PRICE_BELOW_BB_LOWER')
                signals['reasons'].append(f'Giá chạm dải BB dưới ({bb_lower:.1f}) - Oversold')
            elif current_price > bb_upper:
                score -= 1.5
                signals['signals'].append('PRICE_ABOVE_BB_UPPER')
                signals['reasons'].append(f'Giá chạm dải BB trên ({bb_upper:.1f}) - Overbought')
        
        # === Stochastic Analysis ===
        stoch_k = self.indicators['stoch_k'].iloc[-1]
        stoch_d = self.indicators['stoch_d'].iloc[-1]
        
        if not any(np.isnan([stoch_k, stoch_d])):
            if stoch_k < 20 and stoch_d < 20:
                score += 1
                signals['signals'].append('STOCH_OVERSOLD')
                signals['reasons'].append(f'Stochastic oversold ({stoch_k:.1f}) - Tín hiệu mua')
            elif stoch_k > 80 and stoch_d > 80:
                score -= 1
                signals['signals'].append('STOCH_OVERBOUGHT')
                signals['reasons'].append(f'Stochastic overbought ({stoch_k:.1f}) - Tín hiệu bán')
        
        # === ADX (Trend Strength) ===
        adx = self.indicators['adx'].iloc[-1]
        if not np.isnan(adx):
            if adx > 25:
                signals['reasons'].append(f'ADX = {adx:.1f} - Xu hướng mạnh')
            else:
                signals['reasons'].append(f'ADX = {adx:.1f} - Thị trường sideway')
        
        # === Volume Analysis ===
        current_volume = self.data['volume'].iloc[-1]
        avg_volume = self.data['volume'].rolling(20).mean().iloc[-1]
        
        if not np.isnan(avg_volume):
            volume_ratio = current_volume / avg_volume
            if volume_ratio > 1.5:
                signals['reasons'].append(f'Khối lượng tăng {volume_ratio:.1f}x - Dòng tiền mạnh')
                score += 0.5
            elif volume_ratio < 0.5:
                signals['reasons'].append(f'Khối lượng thấp {volume_ratio:.1f}x - Dòng tiền yếu')
                score -= 0.5
        
        # === Final Decision ===
        signals['technical_score'] = score
        
        if score >= 4:
            signals['recommendation'] = 'STRONG BUY'
            signals['confidence'] = min(0.95, 0.7 + (score - 4) * 0.05)
        elif score >= 2:
            signals['recommendation'] = 'BUY'
            signals['confidence'] = min(0.85, 0.6 + (score - 2) * 0.05)
        elif score >= -2:
            signals['recommendation'] = 'HOLD'
            signals['confidence'] = 0.5 + abs(score) * 0.05
        elif score >= -4:
            signals['recommendation'] = 'SELL'
            signals['confidence'] = min(0.85, 0.6 + abs(score + 2) * 0.05)
        else:
            signals['recommendation'] = 'STRONG SELL'
            signals['confidence'] = min(0.95, 0.7 + abs(score + 4) * 0.05)
        
        # === Support & Resistance Levels ===
        signals['support_resistance'] = self._calculate_support_resistance()
        
        # === Target Price ===
        signals['price_targets'] = self._calculate_price_targets()
        
        return signals
    
    def _calculate_support_resistance(self) -> Dict:
        """
        Tính các mức hỗ trợ và kháng cự
        """
        df = self.data.tail(50)  # Lấy 50 phiên gần nhất
        
        # Tìm local minima (support) và local maxima (resistance)
        from scipy.signal import argrelextrema
        
        high_idx = argrelextrema(df['high'].values, np.greater, order=5)[0]
        low_idx = argrelextrema(df['low'].values, np.less, order=5)[0]
        
        resistance_levels = sorted(df.iloc[high_idx]['high'].values, reverse=True)[:3]
        support_levels = sorted(df.iloc[low_idx]['low'].values, reverse=True)[:3]
        
        return {
            'resistance': [float(x) for x in resistance_levels],
            'support': [float(x) for x in support_levels]
        }
    
    def _calculate_price_targets(self) -> Dict:
        """
        Ước tính giá mục tiêu dựa trên Bollinger Bands và Fibonacci
        """
        current_price = self.data['close'].iloc[-1]
        bb_upper = self.indicators['bb_upper'].iloc[-1]
        bb_lower = self.indicators['bb_lower'].iloc[-1]
        sma_20 = self.indicators['sma_20'].iloc[-1]
        
        # Recent high/low
        recent_high = self.data['high'].tail(20).max()
        recent_low = self.data['low'].tail(20).min()
        
        # Fibonacci levels
        diff = recent_high - recent_low
        fib_levels = {
            '0.236': recent_low + diff * 0.236,
            '0.382': recent_low + diff * 0.382,
            '0.500': recent_low + diff * 0.500,
            '0.618': recent_low + diff * 0.618,
            '0.786': recent_low + diff * 0.786
        }
        
        return {
            'current': float(current_price),
            'short_term_target': float(bb_upper),
            'short_term_stop': float(bb_lower),
            'medium_term_target': float(recent_high),
            'medium_term_stop': float(sma_20),
            'fibonacci_levels': {k: float(v) for k, v in fib_levels.items()}
        }
    
    def get_full_analysis(self) -> Dict:
        """
        Trả về phân tích kỹ thuật toàn diện
        """
        # Fetch data nếu chưa có
        if self.data is None:
            self.fetch_data()
        
        # Calculate indicators
        self.calculate_all_indicators()
        
        # Generate signals
        signals = self.generate_signals()
        
        # Compile full report
        report = {
            'symbol': self.symbol,
            'analysis_date': datetime.now().isoformat(),
            'market_data': {
                'current_price': float(self.data['close'].iloc[-1]),
                'open': float(self.data['open'].iloc[-1]),
                'high': float(self.data['high'].iloc[-1]),
                'low': float(self.data['low'].iloc[-1]),
                'volume': int(self.data['volume'].iloc[-1]),
                'change_percent': float((self.data['close'].iloc[-1] - self.data['close'].iloc[-2]) / self.data['close'].iloc[-2] * 100)
            },
            'indicators': {
                'rsi': float(self.indicators['rsi'].iloc[-1]),
                'macd': float(self.indicators['macd'].iloc[-1]),
                'macd_signal': float(self.indicators['macd_signal'].iloc[-1]),
                'sma_20': float(self.indicators['sma_20'].iloc[-1]),
                'sma_50': float(self.indicators['sma_50'].iloc[-1]),
                'bb_upper': float(self.indicators['bb_upper'].iloc[-1]),
                'bb_lower': float(self.indicators['bb_lower'].iloc[-1]),
                'stoch_k': float(self.indicators['stoch_k'].iloc[-1]),
                'adx': float(self.indicators['adx'].iloc[-1])
            },
            'signals': signals,
            'volatility': {
                'atr': float(self.indicators['atr'].iloc[-1]),
                'bb_width': float(self.indicators['bb_width'].iloc[-1])
            }
        }
        
        return report


# === USAGE EXAMPLE ===
if __name__ == "__main__":
    # Test với mã VNM (Vinamilk)
    agent = TechnicalAnalysisAgent(symbol='VNM', days=90)
    
    try:
        analysis = agent.get_full_analysis()
        
        print(f"\n{'='*60}")
        print(f"PHÂN TÍCH KỸ THUẬT: {analysis['symbol']}")
        print(f"{'='*60}\n")
        
        print(f"Giá hiện tại: {analysis['market_data']['current_price']:,.0f} VND")
        print(f"Thay đổi: {analysis['market_data']['change_percent']:+.2f}%")
        print(f"\nKHUYẾN NGHỊ: {analysis['signals']['recommendation']}")
        print(f"Độ tin cậy: {analysis['signals']['confidence']:.1%}")
        print(f"Điểm kỹ thuật: {analysis['signals']['technical_score']}")
        
        print(f"\nCÁC TÍN HIỆU:")
        for reason in analysis['signals']['reasons']:
            print(f"  • {reason}")
        
        print(f"\nMỨC GIÁ MỤC TIÊU:")
        targets = analysis['signals']['price_targets']
        print(f"  • Target ngắn hạn: {targets['short_term_target']:,.0f} VND")
        print(f"  • Stop loss: {targets['short_term_stop']:,.0f} VND")
        
    except Exception as e:
        print(f"Error: {e}")
