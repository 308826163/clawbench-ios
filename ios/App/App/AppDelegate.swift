import UIKit
import WebKit

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {
    var window: UIWindow?
    private var cachedWebView: WKWebView?

    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        // 启动时一次性获取并缓存WKWebView
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            if let mainWindow = application.windows.first,
               let rootVC = mainWindow.rootViewController,
               let webView = rootVC.view.subviews.first(where: { $0 is WKWebView }) as? WKWebView {
                self.cachedWebView = webView
                let scroll = webView.scrollView
                scroll.contentInsetAdjustmentBehavior = .never
                scroll.bounces = false
            }
        }

        // 注册键盘全局监听
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(keyboardWillShow(_:)),
            name: UIResponder.keyboardWillShowNotification,
            object: nil
        )
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(keyboardWillHide(_:)),
            name: UIResponder.keyboardWillHideNotification,
            object: nil
        )
        return true
    }

    @objc private func keyboardWillShow(_ notify: Notification) {
        guard let webView = cachedWebView else { return }
        guard let kbFrame = notify.userInfo?[UIResponder.keyboardFrameEndUserInfoKey] as? CGRect else { return }
        let kbHeight = kbFrame.height
        let animDuration: TimeInterval = 0.28
        UIView.animate(withDuration: animDuration, delay: 0, options: .curveEaseOut) {
            webView.scrollView.contentInset = UIEdgeInsets(top: 0, left: 0, bottom: kbHeight, right: 0)
        }
    }

    @objc private func keyboardWillHide(_ notify: Notification) {
        guard let webView = cachedWebView else { return }
        let animDuration: TimeInterval = 0.28
        UIView.animate(withDuration: animDuration, delay: 0, options: .curveEaseOut) {
            webView.scrollView.contentInset = .zero
        }
    }

    deinit {
        NotificationCenter.default.removeObserver(self)
    }

    func applicationDidEnterBackground(_ application: UIApplication) {}
    func applicationWillEnterForeground(_ application: UIApplication) {}
}
